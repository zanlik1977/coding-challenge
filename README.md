# Golang Coding challenge

Write an application for fetching content from multiple providers in a simple http service that simulates a news API.

## My approach

I have decided to approach this challenge from the point of trying to keep the 
original service unchanged as much as possible, and follow the instructions fully:

1. Complete the `ServeHTTP` method in server.go in accordance with the specifications above.

    I have completed `ServeHTTP` method in server.go :

    - Parse the request for count and offset.
    - Set the list of Providers (either main or fallback) to use, in the repeating order of the default configuration start from offset if there is one.
   
    - Then I make requests for main Providers.
    - Launch goroutines in the predetermined order for Providers. Since the execution time varies for each goroutine items will not be returned in predetermined order. Index variable keeps this order in place.
    - Whichever item is returned through the channel, it will be matched to the appropriate Provider in providerListToUse.
    - Whichever goroutime finishes first, item goes to channel
    - As other goroutines finish, items they return wait to get into the channel one by one. Only one item gets out of channel at the time.
    - Sending response (content item) through unbuffered channel keeps threads safe as only a single goroutine can access a single value in the channel. (This can also be achieved with buffered channel but I did not have time to play with it).

    - Next, make requests for fallback Providers.
    - If there are failures in the Items map (empty content returned), use the list of fallback Providers to make a new request for the fallback Provider for each faillure.

    - At the end, use processedForDoubleFailure to check if both the main provider and the fallback fail (or if the main provider fails and there is no fallback), and return the items before that point.

    - One more thing. I have implemented ResponseWritter.


2. Run existing tests, and make sure they all pass.

    I made a minor change to runRequest as I'm returning map of Items instead of slice.
    All existing tests are passing.

3. Add a few tests to capture missing edge-cases. For example, test that the fallbacks are respected.

    Added test for request "GET", "/?offset=10&count=100"
    Added tests for testing fallback:

        Set of tests where I force fallback on Provider3:
            - TestResponseCountWithFallbackProvider3
            - TestResponseOrderWithFallbackProvider3

        Set of tests where main Provider fails but there is no fallback. I force fallback on Provider1 which leaves config4 with no fallback:
            - TestResponseCountMainFailsNoFallback
            - TestResponseOrderMainFailsNoFallback

        Set of tests where main fails and fallback fails. I force fallback on Provider2 and Provider3:
            - TestResponseCountMainFailsFallbackFails
            - TestResponseOrderMainFailsFallbackFails


## Run the server and tests

    You can run the server simply with `go run .` in the projects directory.
    You can run the tests simply with `go test` in the projects directory.
    
    
## Final note

    Overall, I have enjoyed working on this challenge.
    Also, I'm aware I could have done a better job with error handling and perhaps some methods could be broken into more.....
