# ok (overkill) is just a test playground for the time being.

# Build
``` bash 
$ git clone https://github.com/mclellac/ok 
$ cd ok
$ make deps && make proto && make
$ make install
```

# Start MariaDB/MySQL server and create the DB
``` sql
mysql> create database posts;
```

# postd configuration file
Copy the example configuration file from ./servers/posts, and modify what you need to.

# Start the postd server
``` bash 
$ postd
```

# Add a post with the client
``` bash
$ ok add "I'm a test post title" "I'm the body of the post."
```
