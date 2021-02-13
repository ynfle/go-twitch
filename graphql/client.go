package graphql

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Adeithe/go-twitch/api"
	"github.com/shurcooL/graphql"
)

// Client stores data about a GraphQL client
type Client struct {
	ID     string
	bearer string

	graphql *graphql.Client
}

// URL is the address for the GraphQL server
const URL = "https://gql.twitch.tv/gql"

// New Twitch GraphQL Client
//
// This uses the official Twitch client by default and therefore should be used sparingly or not at all.
func New() (client *Client) {
	client = &Client{ID: api.Official.ID}
	client.graphql = graphql.NewClient(URL, &http.Client{Transport: httpTransport{client, http.DefaultTransport}})
	return
}

// SetBearer sets the token sent with GraphQL requests
func (client *Client) SetBearer(token string) {
	client.bearer = token
}

// CustomQuery executes a query on the GraphQL server
//
// See: https://github.com/shurcooL/graphql
func (client Client) CustomQuery(query interface{}, vars map[string]interface{}) error {
	return client.graphql.Query(context.Background(), query, vars)
}

// CustomMutation executes a mutation on the GraphQL server
//
// See: https://github.com/shurcooL/graphql
func (client Client) CustomMutation(mutation interface{}, vars map[string]interface{}) error {
	return client.graphql.Mutate(context.Background(), mutation, vars)
}

// IsUsernameAvailable returns true if the provided username is not taken on Twitch
func (client *Client) IsUsernameAvailable(username string) (bool, error) {
	query := GQLUsernameAvailabilityQuery{}
	vars := map[string]interface{}{"username": graphql.String(username)}
	err := client.CustomQuery(&query, vars)
	return query.IsAvailable, err
}

// GetCurrentUser retrieves the current user based on the clients authentication token
func (client Client) GetCurrentUser() (*User, error) {
	if len(client.bearer) < 1 {
		return nil, ErrTokenNotSet
	}
	query := GQLCurrentUserQuery{}
	err := client.CustomQuery(&query, nil)
	return query.Data, err
}

// GetUsersByID retrieves an array of users from Twitch based on their User IDs
func (client Client) GetUsersByID(ids ...string) ([]User, error) {
	if len(ids) > 100 {
		return []User{}, ErrTooManyArguments
	}
	query := GQLUserIDsQuery{}
	vars := map[string]interface{}{"ids": toIDs(ids...)}
	err := client.CustomQuery(&query, vars)
	return query.Data, err
}

// GetUsersByLogin retrieves an array of users from Twitch based on their usernames
func (client Client) GetUsersByLogin(logins ...string) ([]User, error) {
	if len(logins) > 100 {
		return []User{}, ErrTooManyArguments
	}
	query := GQLUserLoginsQuery{}
	vars := map[string]interface{}{"logins": toStrings(logins...)}
	err := client.CustomQuery(&query, vars)
	return query.Data, err
}

// GetChannelsByID retrieves an array of channels from Twitch based on their IDs
func (client Client) GetChannelsByID(ids ...string) ([]Channel, error) {
	if len(ids) > 100 {
		return []Channel{}, ErrTooManyArguments
	}
	query := GQLChannelIDsQuery{}
	vars := map[string]interface{}{"ids": toIDs(ids...)}
	err := client.CustomQuery(&query, vars)
	return query.Data, err
}

// GetChannelsByName retrieves an array of channels from Twitch based on their names
func (client Client) GetChannelsByName(names ...string) ([]Channel, error) {
	if len(names) > 100 {
		return []Channel{}, ErrTooManyArguments
	}
	query := GQLChannelNamesQuery{}
	vars := map[string]interface{}{"names": toStrings(names...)}
	err := client.CustomQuery(&query, vars)
	return query.Data, err
}

// GetStreams retrieves data about streams available on Twitch
func (client Client) GetStreams(opts StreamQueryOpts) (*StreamsQuery, error) {
	if opts.First < 1 || opts.First > 100 {
		opts.First = 25
	}
	query := GQLStreamsQuery{}
	vars := map[string]interface{}{
		"first":   graphql.Int(opts.First),
		"after":   opts.After,
		"options": opts.Options,
	}
	err := client.CustomQuery(&query, vars)
	return query.Data, err
}

// GetVideos retrieves videos on Twitch
func (client Client) GetVideos(opts VideoQueryOpts) (*VideosQuery, error) {
	if opts.First < 1 || opts.First > 100 {
		opts.First = 25
	}
	query := GQLVideosQuery{}
	vars := map[string]interface{}{
		"first": graphql.Int(opts.First),
		"after": opts.After,
	}
	err := client.CustomQuery(&query, vars)
	return query.Data, err
}

// GetVideosByChannel retrieves videos on Twitch based on the provided channel
func (client Client) GetVideosByChannel(channel Channel, opts VideoQueryOpts) (*VideosQuery, error) {
	return client.GetVideosByUser(User{ID: channel.ID}, opts)
}

// GetVideosByUser retrieves videos on Twitch based on the provided user
func (client Client) GetVideosByUser(user User, opts VideoQueryOpts) (*VideosQuery, error) {
	if opts.First < 1 || opts.First > 100 {
		opts.First = 25
	}
	query := GQLUserVideosQuery{}
	vars := map[string]interface{}{
		"id":    user.ID,
		"first": graphql.Int(opts.First),
		"after": opts.After,
	}
	err := client.CustomQuery(&query, vars)
	if query.Data == nil {
		return nil, err
	}
	return query.Data.Videos, err
}

// GetClipBySlug retrieves data about a clip available on Twitch by its slug
func (client Client) GetClipBySlug(slug string) (*Clip, error) {
	query := GQLClipQuery{}
	vars := map[string]interface{}{"slug": slug}
	err := client.CustomQuery(&query, vars)
	return query.Data, err
}

// GetGames retrieves data about games available on Twitch
func (client Client) GetGames(opts GameQueryOpts) (*GamesQuery, error) {
	if opts.First < 1 || opts.First > 100 {
		opts.First = 25
	}
	query := GQLGamesQuery{}
	vars := map[string]interface{}{
		"first":   graphql.Int(opts.First),
		"after":   opts.After,
		"options": opts.Options,
	}
	err := client.CustomQuery(&query, vars)
	return query.Data, err
}

// GetFollowersForUser retrieves data about who follows the provided user on Twitch
func (client Client) GetFollowersForUser(user User, opts FollowQueryOpts) (*FollowersQuery, error) {
	if user.ID == nil || len(fmt.Sprint(user.ID)) < 1 {
		return nil, ErrInvalidArgument
	}
	if opts.First < 1 || opts.First > 100 {
		opts.First = 25
	}
	query := GQLFollowersQuery{}
	vars := map[string]interface{}{
		"id":    user.ID,
		"first": graphql.Int(opts.First),
		"after": opts.After,
	}
	err := client.CustomQuery(&query, vars)
	if query.Data == nil {
		return nil, err
	}
	return query.Data.Followers, err
}

// GetFollowersForChannel retrieves data about who follows the provided channel on Twitch
func (client Client) GetFollowersForChannel(channel Channel, opts FollowQueryOpts) (*FollowersQuery, error) {
	if channel.ID == nil || len(fmt.Sprint(channel.ID)) < 1 {
		return nil, ErrInvalidArgument
	}
	if opts.First < 1 || opts.First > 100 {
		opts.First = 25
	}
	query := GQLFollowersQuery{}
	vars := map[string]interface{}{
		"id":    channel.ID,
		"first": graphql.Int(opts.First),
		"after": opts.After,
	}
	err := client.CustomQuery(&query, vars)
	if query.Data == nil {
		return nil, err
	}
	return query.Data.Followers, err
}

// GetModsForChannel retrieves data about who is a moderator for the provided channel on Twitch
func (client Client) GetModsForChannel(channel Channel, opts ModsQueryOpts) (*ModsQuery, error) {
	return client.GetModsForUser(User{ID: channel.ID}, opts)
}

// GetVIPsForChannel retrieves data about who is a VIP for the provided channel on Twitch
func (client Client) GetVIPsForChannel(channel Channel, opts VIPsQueryOpts) (*VIPsQuery, error) {
	return client.GetVIPsForUser(User{ID: channel.ID}, opts)
}

// GetModsForUser retrieves data about who is a moderator for the provided user on Twitch
func (client Client) GetModsForUser(user User, opts ModsQueryOpts) (*ModsQuery, error) {
	if user.ID == nil || len(fmt.Sprint(user.ID)) < 1 {
		return nil, ErrInvalidArgument
	}
	if opts.First < 1 || opts.First > 100 {
		opts.First = 25
	}
	query := GQLModsQuery{}
	vars := map[string]interface{}{
		"id":    user.ID,
		"first": graphql.Int(opts.First),
		"after": opts.After,
	}
	err := client.CustomQuery(&query, vars)
	return query.Data.Mods, err
}

// GetVIPsForUser retrieves data about who is a VIP for the provided user on Twitch
func (client Client) GetVIPsForUser(user User, opts VIPsQueryOpts) (*VIPsQuery, error) {
	if user.ID == nil || len(fmt.Sprint(user.ID)) < 1 {
		return nil, ErrInvalidArgument
	}
	if opts.First < 1 || opts.First > 100 {
		opts.First = 25
	}
	query := GQLVIPsQuery{}
	vars := map[string]interface{}{
		"id":    user.ID,
		"first": graphql.Int(opts.First),
		"after": opts.After,
	}
	err := client.CustomQuery(&query, vars)
	return query.Data.VIPs, err
}
