Go MMO Backend Thing
====

After talking to a junior dev about a side project they were working on, wanted
to give the idea a shot.

## The Idea

A backend server for a simple MMO.

The game is pretty simple:

* The game takes place on a 50x50 grid
* Each square can contain only one player
* Each square has a depth, up to a maximum depth of 10 units
* Each square may contain treasure
* To find treasure, player must dig down until they reach a level with treasure

## The Server

Does two things:

1. Keeps track of the state of the world, and sends it out once a second to all
   connected clients
2. Accepts incoming messages from the player client to update the state of the world   

## The Client

Sends messages to the server to update the state of the game.

## Messages

The message format is like so:

```
{
  "type": "<type string>"
}
```

Where type string is one of:

* connect
* up
* down
* left
* right
* dig
* descend
* climb

