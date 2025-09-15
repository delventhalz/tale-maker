# Tale Maker

The goal of this project is to create a language for non-technical writers to
describe a complete interactive text adventure game. It is an outgrowth of the
board game [Taelmoor](https://taelmoor.com) and may one day be used to write
Taelmoor scenarios. This repo will also eventually host a minimal text adventure
engine which can be used to play written tales without the Taelmoor board game.

## How it works

Text will be organized into locations and nested actions such as "examine" or
"press". Actions can be nested within each other and game objects can be treated
as actions. This "examine" and "barrel" nested within the location "entrance",
would include the text for examining the barrel located in the entrance.

Text will also be able to include commands, which can do things like set
variables, change text style, or trigger an effect in the engine. There will be
both engine-specific and basic Tale Maker commands. If an engine running a tale
does not recognize a command, it is ignored.

Tales can be written over multiple files and combined arbitrarily. Order of
location blocks and any config/setup does not matter. Config steps can executed
only once and will cause an error if run twice.

## Whats next

I will begin by reproducing Taelmoor's tutorial scenario "Initiation Test" as a
tale, and running it in a CLI based text adventure engine. As I dogfood my own
syntax and work out the kinks, I will begin documenting it more officially, and
eventually arrive at some alpha/beta versions of the language.
