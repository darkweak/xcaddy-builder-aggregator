package aggregator

import "github.com/shurcooL/githubv4"

type LatestRelease struct {
	TagName   githubv4.String
	TagCommit struct {
		Commit struct {
			Tree struct {
				Entries []struct {
					Name   githubv4.String
					Path   githubv4.String
					Object struct {
						Tree struct {
							Entries []struct {
								Name   githubv4.String
								Path   githubv4.String
								Object struct {
									Tree struct {
										Entries []struct {
											Name   githubv4.String
											Path   githubv4.String
											Object struct {
												Tree struct {
													Entries []struct {
														Name githubv4.String
														Path githubv4.String
													} `graphql:"entries"`
												} `graphql:"... on Tree"`
											} `graphql:"object"`
										} `graphql:"entries"`
									} `graphql:"... on Tree"`
								} `graphql:"object"`
							} `graphql:"entries"`
						} `graphql:"... on Tree"`
					} `graphql:"object"`
				} `graphql:"entries"`
			} `graphql:"tree"`
		} `graphql:"... on Commit"`
		CommitUrl githubv4.String
	} `graphql:"tagCommit"`
}

type GithubRepositoryRetriever struct {
	Search struct {
		Edges []struct {
			Node struct {
				Repository struct {
					Url           githubv4.String
					Description   githubv4.String
					NameWithOwner githubv4.String
					RepositoryTopics struct {
						Nodes []struct {
							Topic struct {
								Name githubv4.String
							} `graphql:"topic"`
						} `graphql:"nodes"`
					} `graphql:"repositoryTopics(first: 100)"`
					LatestRelease LatestRelease `graphql:"latestRelease"`
				} `graphql:"... on Repository"`
			} `graphql:"node"`
		} `graphql:"edges"`
	} `graphql:"search(query: \"topic:caddy-module sort:updated-desc\", type: REPOSITORY, first: 100)"`
}

func newRepositoryRetriever() GithubRepositoryRetriever {
	return GithubRepositoryRetriever{}
}
