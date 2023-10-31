# vuLnDAP

vuLnDAP is a deliberately vulnerable web application to demonstrate exploiting business logic flaws in a site based on LDAP.

For more information see the [project homepage](https://digi.ninja/projects/vulndap.php).

If you get stuck and need help, I've written a [walkthrough](https://digi.ninja/blog/vulndap_walkthrough.php).

If you want to have a play, but do not want to install it yourself, you can use [my copy](https://vulndap.digi.ninja/). Please do not abuse this, the vulnerabilities are in the application, do not expect to get anywhere with port scans or by running Nessus against it.

## Installation

```
go get github.com/digininja/vuLnDAP
```

Change into the vuLnDAP directory which will probably be ~/go/src/github.com/digininja/vuLnDAP then run the following command to install the dependencies:

```
go get -d ./...
```

You can then build the app with:

```
go build
```

And if all goes well you'll get a vuLnDAP binary which you can run with:

```
./vuLnDAP
```

The web server starts up listening on port 9090 on all interfaces, this can be changed by updating the values in the webserver section of vulndap.cfg.



## Building from Source

```
git clone https://github.com/digininja/vuLnDAP.git

cd vuLnDAP
```

Clone the repository, then change into that directory

```
go get 
```

Download all the necessary 3rd Party Packages


```
go build
```

Build all the files to create a `./vuLnDAP` binary


### Configuration

```
cp vulndap.cfg-sample vulndap.cfg
```

Prior to Usage, ensure that a `vulndap.cfg` file exists. For default configurations, refer to `vulndap.cfg-sample`


### Usage 

In order to run, ensure that you have successfully built a `./vuLnDAP` binary. If so simply run:

```
./vuLnDAP
```

The web server starts up listening on port 9090 on all interfaces, this can be changed by updating the values in the webserver section of vulndap.cfg.
