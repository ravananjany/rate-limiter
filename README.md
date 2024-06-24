# limiter middleware 
     this rate limiter application works on token bucket algorithm.
     we have used golang inbuilt rate/limiter package as base and we have created our own custom dynamically configurable rate limiter.  
     go version 1.21.
     we have unit tested the limiter function.


# Usage:
     using this package we can create limiters for clients and endpoints.
     DynamicConfigurableLimiter interface provides access to configure rate limiter 
     config json file containing client data is used create client limiters dynamically on startup (ex - config file is in the source dir)
     we can configurae limiter by providing respective json data in the config file (limts, bursts)

# enchancement:
     we can have listeners(channels) to update client data in limiter middleware from the application that used this middleware
     we can have a functionality that created limiters based on data recvd from listeners 
     
