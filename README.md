actor
=====

                _             
               | |            
      __ _  ___| |_ ___  _ __ 
     / _` |/ __| __/ _ \| '__|
    | (_| | (__| || (_) | |   
     \__,_|\___|\__\___/|_|   
    
### Install

    # install golang amd64 
    open https://golang.org/dl/

    # setup GOPATH env
    export GOPATH=~/gopkg

    # install actor package
    go get github.com/funkygao/dragon

    # change dir to actor package
    cd $GOPATH/src/github.com/funkygao/dragon

    # install all dependencies
    go get ./... 

    # build the executable
    ./build.sh

    # create a custom config file based upon sample
    cp etc/actord.cf.sample etc/actord.cf (change the address to your own)

    # startup the daemon
    ./daemon/actord/actord

### actor IS

* external scheduler for delayed jobs
  - PvP march
  - PvE march
  - Job

* serializer for concurrent updates
  - lock container/issuer with retry mechanism
  - actor make concurrent calls into sequential calls

* coodinator
  - everything that may lead to race condition and concurrent updates will be decided by actor

### TODO

* worker php, retry only when errno=1(locked)
* when faed restart, actor couldn't get lock any more. try restart fae, what actor happens?
* what if beanstalk conn broken?
* A encamp，然后1万人同时到，轮流打他，那么这个阶段，A啥也干不了。城外人的锁应该特殊处理
* mysql ping within mysql max idle time, my.cf wait_timeout show variables like '%timeout%'
* metrics of php request handling
* test callback php timeout
* test lock expiry
* March may need K
* alliance lock
* pprof may influnce performance
* mysql transaction with isolation repeatable read + optimistic locking has same effect
  - I'd rather kill actor instead of mysqld
  - what about distributed mysql instances?
* teleport
*   Write/Read timeout and check N in loop
*   can a player send N marches to the same tile?
*   simulate mysql shutdown
    - done! golang mysql driver with breaker will handle this
*   WHERE UNIX_TIMESTAMP(time_end) index hit
    - need to optimize DB index
*   worker throttle
    - we can't have toooo many callbacks concurrently, use channel for throttle, easy...
*   handles NULL column
    - march.type done if it's NULL, what about others?
    - maybe we should let DB handle this
    - but mysql enum datatype can't handle this automatically
*   tsung 20M rows in db, and try actor

