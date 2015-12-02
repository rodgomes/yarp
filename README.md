# yarp
Yet another reverse proxy in Golang

Just an exercise with something that I need, because I do need a simple proxy to run some stuff locally, but I don't want to install and setup anything relativelly complex. 
BTW, I still need a clean up and adding some tests.

To configure it, you just have to create a json file with the following structure and give it in the command line with "-c" argument.

{
  "port": "9090", //this is the port the proxy itself will run
  //then you provide a list of routers, which may or may not contain a pattern to match

  "routers" : [
    {
      "targetUrl": "localhost:8001",  //the url:port of target host
      "scheme": "http",               //well, the scheme
      "pathPattern": "/some/path/.*?" //optional regex pattern to route to this target. If not present (like router bellow)
                                      //will route all incoming requests
    },
    {
      "targetUrl": "localhost:8000",
      "scheme": "http"
    }
  ]
}