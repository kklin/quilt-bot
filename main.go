package main

import (
	"fmt"
	"golang.org/x/net/context"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

type starGazer struct {
	username  string
	avatarURL string
}

func main() {
	fmt.Println("Started!")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_OATH")},
	)
	tc := oauth2.NewClient(ctx(), ts)

	client := github.NewClient(tc)

	var users []starGazer
	for range time.Tick(15 * time.Second) {
		new, err := getUsers(client)
		if err != nil {
			fmt.Println("Failed to get stargazers: ", err)
			continue
		}

		if users != nil {
			runStarCheck(users, new)
		}
		users = new

		runReview(client)
	}
}

func runStarCheck(users, new []starGazer) {
	additions := extraStargazer(users, new)
	deletions := extraStargazer(new, users)

	total := len(new)

	for _, add := range additions {
		fmt.Println("New Star!", add.username)
		post(add, total, true)
	}

	for _, del := range deletions {
		fmt.Println("Lost a Star!", del.username)
		post(del, total, false)
	}
}

func post(user starGazer, total int, add bool) {
	iconemoji := ":confetti_ball:"
	color := "#009900" // Green
	if !add {
		iconemoji = ":slightly_frowning_face:"
		color = "#D00000" // Red
	}

	title := "We've got a new star!"
	if !add {
		title = "We've lost a star!"
	}

	un := user.username
	text := fmt.Sprintf(
		"<https://github.com/NetSys/quilt/stargazers|%d Quilt Stargazers>\n\n",
		total)
	text += fmt.Sprintf("<https://github.com/%s|%s>", un, un)

	slack(os.Getenv("SLACK_ENDPOINT"), slackPost{
		Channel:   os.Getenv("SLACK_CHANNEL"),
		Color:     color,
		Username:  "quilt-bot",
		Iconemoji: iconemoji,
		Fields: []message{
			{
				Title: title,
				Short: false,
				Value: text,
			},
		},
	})
}

// Returns all stargazers in `check` that are not in `base`.
func extraStargazer(base, check []starGazer) []starGazer {
	baseMap := map[string]struct{}{}
	for _, b := range base {
		baseMap[b.username] = struct{}{}
	}

	var extras []starGazer
	for _, c := range check {
		if _, ok := baseMap[c.username]; !ok {
			extras = append(extras, c)
		}
	}
	return extras
}

func getUsers(client *github.Client) ([]starGazer, error) {
	var results []starGazer

	opt := &github.ListOptions{}
	for {
		sgs, resp, err := client.Activity.ListStargazers(ctx(), "quilt", "quilt", opt)
		if err != nil {
			return nil, err
		}

		for _, s := range sgs {
			results = append(results, starGazer{
				username:  *s.User.Login,
				avatarURL: *s.User.AvatarURL,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return results, nil
}

func ctx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx
}
