# greenlight

Global Movies API

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
  and to interact with database i used the DATABASE/SQL package, it's very fast & strightforward, manual mapping SQL fileds to variable, of course there are other options to interact with the postgres such as GORM( it's a CRUD fucntion already implemented, very short production code the, cons: run slowly on high load, ... run slower 6x time than DATABASE/SQL), and SQLC(it's very fast & easy to use, automatic code generation, catch errors before generating code).

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
