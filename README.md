# httpproxy
<h1>a homework to make a httpproxy server with logging sites and make a list of categories of sites and see how much you spent in them </h1>
After running main.go http proxy server start at port 8080 and a simple server to serve and change configs and api’s. 

Front of table to show data is designed with angular in dist folder. 

API : 

Localhost:4300 - return the table of how much you requested to sites in your category list 

Localhost:4300/info - api for angular to show data (json of data in previous api) 

Localhost:4300/read_config-file - return configs 

Localhost:4300/write_config-file - rewrite config file 
