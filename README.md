# Tale Maker

The goal of this project is to create a language for non-technical writers to
describe an interactive text adventure game. It is an outgrowth of the board
game [Taelmoor](https://taelmoor.com), and may one day be used to write Taelmoor
scenarios. This repo will eventually host a command line tool which can play
Tale Maker games directly, export a JSON format suitable for browser-based game
engines to play, as well as export the original Taelmoor scenario format.

## How it works

Text is organized into blocks and each block is linked to particular inputs from
the player and a particular game state. These conditions allow Tale Maker to pick
a single block of text to display after each player input.

In order to manipulate game state, tale authors can insert actions into text
blocks, changing variables, moving the player, or styling text. In this way, an
interactive narrative can unfold, one player input at a time.

For more, checkout the high level [overview](./docs/overview.md) of the Tale
Maker syntax.

## Whats next

The first step is building out a complete Tale Maker parser to run tale files
and play games in the CLI. With that working, I will shift focus to writing and
playing tales, including reproducing some existing Taelmoor scenarios. Once I am
happy with the syntax, JSON formats to move Tale Maker out of the CLI and into
the browser (and perhaps the Taelmoor app) will be my next focus.
