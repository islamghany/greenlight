# Greenlight

Global Movies API

[Greenlight api Docs](https://app.swaggerhub.com/apis-docs/islamghany1/greenlight/1.0.0)

## Installation

The installation is very simple the only required thing is to have docker and docker compose installed on your machine.

just use the command.

```bash
docker-compose up
```

to stop the containers hit _Ctrl+C_

## How things works

### database

- The database that was selected for this project was PosgresSQL. <br/>
  and to interact with database i used the DATABASE/SQL package, it's very fast & straightforward, manual mapping SQL fields to variable, of course there are other options to interact with the postgres such as GORM( it's a CRUD function already implemented, very short production code the, cons: run slowly on high load, ... run slower 6x time than DATABASE/SQL), and SQLC(it's very fast & easy to use, automatic code generation, catch errors before generating code).

- Establishing the connection pool:<br/>
  estimating the ‘in-use’ connections and ‘idle’ connections.<br/>
  for this project i have set the max open connections to 25 connections. I’ve found this to be a reasonable starting point for small-to-medium web applications and APIs.
  and for better performance i have set max idle connections to the same number i used in the max open connections, and for memory sake to remove the idle connections that aren't used for a long time i have set the max idle time duration to 15 minutes, It’s probably OK to leave max open connections as unlimited, unless your database imposes a hard limit on connection lifetime, or you need it specifically to facilitate something like gracefully swapping databases. Neither of those things apply in this project.

- Concurrency Control:
  to get around the problem of data race there were many options but for me i choosed from two options.<br/>
  - Shared & Exclusive locks, where i can lock a row or a column until i finish my query or my transaction, but this is a solution that i can make in the database itself not a programmatic way.
  - optimistic locking pattern(versioning):<br/>
    The fix works like this:<br/> - Alice and Bob’s goroutines both call GetMovie() to retrieve a copy of the movie record. Both of these records have the version number N. - Alice and Bob’s goroutines make their respective changes to the movie record. - Alice and Bob’s goroutines call UpdateMovie() with their copies of the movie record. But the update is only executed if the version number in the database is still N. If it has changed, then we don’t execute the update and send the client an error message instead.

    This means that the first update request that reaches our database will succeed, and whoever is making the second update will receive an error message instead of having their change applied.
    
### Caching
for cashing i used Redis to store the frequently accessed movies.
i used 80-20 rule, i.e 20% of daily read volume for movies is generating 80% of traffic which means that certain movies are so popular that the majority of users read them, This dictates that we can try caching 20% of daily read volume of movies.

### Security
when designing an app there are some security considerations you must give attention to, here i will demonstrate how i dealt with some of them.

- *IP-based Rate Limiting*: this application is a public api that any one can send requests to it, so it's very easy for the client to making too many requests too quickly and putting excessive strain on the server, and the server could crash.<br/>
<br/>
The solution for this problem is straightforward, create an in-memory map of rate limiters(object that store information about the client IP), using the IP address for each client as the map key.<br/>
when the client sends a request we see the number of requests he made in the last second if there have been too many then it should send the client a `429 Too Many Requests` response.
    *i have set it to 2 requests per second, with a maximum of 4 requests in a burst.*
    <br/>
    there is one problem with the above approach is that the map will grow indefinitely, taking up more and more resources with every new IP address and rate limiter that we add.<br />
    *so to prevent this from happening we can run a function in the background periodically(every 1 minute) that remove  any clients that we haven’t been seen recently from the map(3 minutes).*
    
- *Restricting Inputs:* to get around the problems of SQL Injection and denial-of-service attack, i have to be known what is the data that user give me and what data i want for this specific endpoint and if the provided data suits that endpoint or not.<br />
Things to put in mind when processing the client's provided data
	- the body must be JSON, not malformed, contains no errors.
	- json types must match the types we are trying to decode into.
	- the request must contains body if it's required, *one body*.
	- disallow unknown fields .i.e the client provide an extra field that is not required for this specific endpoint.
	- limit the size of the request body (1 MB)
	- validating JSON Input.
    
   all these challenges i figured out a native solution for them.
   
 ### Autentication & Authorization
  i choose to work with the Stateful token authentication (secure random string. This token — or a fast hash of it — is stored server-side in a database, alongside the user ID and an expiry time for the token.)<br/>
  Stateful authentication tokens is very good for monolithic backend apps.<br />
  although there are a lot of other techniques for authenticating a user, from my point of view this is the best technique suit this app.
   <br /><br />
    Permission-based Authorization: only users who have a specific permission can perform specific operations, for example if user want read from movies he has to have the permission `movies:read` and if he want to write to a movie he has to have `movies:write` permission etc..
    
###  Miscellaneous
there are some functionalities have to exist in any robust app such..<br/>
- Logging system : when an error occurs there must be explanation why error occurs and where it occurs.
- Graceful Shutdown : when the app crash or forcing to shutdown, we must to safely stop the running application.
- Sending mails : when the user register for the first time, link will send to his mail so he can activate his account.
- Metrics : `/debug/vars` Display application metrics .i.e
 performing and what resources it is using.
    
    
    
    
    
    
    
    
