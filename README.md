# 🍽️ DisneyTables

DisneyTables is a program written in Go that retrieves the availability of restaurants in the parks and hotels of Disneyland Paris and tracks the changes in availability of their tables. When a table becomes available according to certain criteria defined by the users, a notification is issued to inform them.
Take a look at [Remy](https://github.com/Romitou/Remy), an example of a Discord bot implementing this project. :)

## Environment variables
* `API_KEY` : your API key to access Disney API services
* `AVAILABILITIES_ENDPOINT` : the Disney API address to retrieve restaurant availabilities
* `GRAPHQL_ENDPOINT` : the Disneyland Paris GraphQL endpoint
* `REFRESH_AUTH_ENDPOINT` : the Disney Auth address to refresh a token access
* `CUSTOM_HEADERS` : a JSON value to integrate specific headers for requests to Disney services
* `RESTAURANTS_QUERY` : the GraphQL query to retrieve restaurants
* `MYSQL_DSN` : the MySQL database connection string
* `REDIS_HOST` : well, the Redis database connection string
* `WEBSERVER_TOKEN` : the external API token to communicate with DisneyTables
* `SENTRY_DSN` : the DSN address to your Sentry configuration

## FAQ

#### How did this idea come about?

This idea came about because of personal difficulties in finding an available table in the restaurants of Disneyland Paris. As many other visitors have the same difficulties with the high demand of the restaurants, I found it useful to develop this application to allow us to get our tables as easily as possible, instead of checking manually when we have a break if a table is available via the application.

#### How does it work in practice?

In practice, the program will check at very regular intervals (configurable) through a request to the Disney API the availability of a notification previously registered by a user. This means that only the necessary calls are made to the Disney API and they are made in a moderate way and constitute a negligible and non-disruptive load on the Disney API.

#### What is the verification interval of a notification?

it depends! Indeed, the process of retrieving availability data is limited in time. This means that within one minute, 5 of the oldest notifications will be checked (or more if customized). The time interval between each check for a specific notification depends on the number of notifications and the "throughput" of the notification check.

#### Can I host this on my end?

Yes and no. Some of the environment variables necessary for DisneyTables to function properly require specific data that require complex technical manipulations. The use of this program is recommended only to informed people: no support for the retrieval of these data will be provided.


