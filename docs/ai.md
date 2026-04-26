## Use of AI

For this project I used Google's Gemini for some basic tasks. I utilized it for the following:

* Generating simple unit test cases after I supplied the names of each case
* Asking general questions about go idioms (the `val, ok := ...` idiom always escapes me)
* Asking about the http library itself

For the unit tests, I ran them all and made sure they made sense. Initially it tried to force me to use a couple of return codes that were not what the requirements specified. Specifically, it wanted to use 404 for some of the object not found returns when the spec asked for 400. It gave me good information about the functionality of the Go HTTP standard library, specifically about extracting the request body, which I always find to be convoluted. 

I found its answers to be useful, and the generated code to be ok as long as I went over it first. It did present to me using a `ServerMux` for test setup, which I hadn't though of in the actual code, so I specifically left mine using the Default Server Mux that the http library uses if you don't specify one. They both work, but perhaps using a custom `ServerMux` is nicer.