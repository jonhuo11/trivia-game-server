# Trivia game server

The game server for my Trivia game. 

This was mainly developed as a test to learn more about concurrency in Go and websockets. I've archived this repo now for a few reasons:
1) Working directly with websockets is difficult. I have to check message type and format on each end to make sure nothing breaks, which ends up being a lot of code just to build something small.
2) I'm no longer interested in working with React + Websockets. I'm shifting my focus to backend projects since that is what I am more interested in now.

If I were to build this again, the main change is I would use gRPC instead of websockets. It is so much easier to work with structured messages provided by gRPC than having to build my own message type and structure enforcement on top of raw websockets.
