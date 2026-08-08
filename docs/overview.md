# Tale Maker Overview

A high level summary of the Tale Maker syntax.

## Blocks

Text is organized into blocks which start with a header and continue until the next block or the end of the file. Text in a block is displayed for the player when the conditions specified in the header are met. Line breaks, indentation, and other white space in the block are typically included in the displayed text, with the exception of empty lines at the start or end of the block. These leading and trailing empty lines are considered padding and ignored.

### Input Block ###

```
> greet >
Hail, Adventurer!
```

The header for an input block starts with one or more `>` characters, is followed by a command input by the player and optionally ends with a matching number of `>` characters.

Multiple `>` characters are used to denote nested blocks. When blocks are nested, players must satisfy both the outer and inner block conditions to trigger the inner block. Text in the outer block is a fallback, displaying when no inner block conditions are met.

```
> greet >
Hail, Adventurer!

>> brigand >>
The brigand eyes you warily. They have nothing to say to you.
```

### State Block ###

```
= room is dark =
It's too dark! You'll need to find a light source.
```

The header for a state block starts with one or more `=` characters, is followed by a condition, and optionally ends with a matching number of `=` characters.

Multiple `=` characters are used to denote nested blocks. Players must satisfy both outer and inner block conditions to trigger the inner block. Text in the outer block is a fallback, displaying when no inner block conditions are met. Input blocks can be nested inside state blocks and vice versa.

```
= room is dark =
It's too dark! You'll need to find a light source.

>> light >>
>>> torch >>>
The torch blazes to life. You can see!
```

### Start Block

The beginning of a Tale Maker file, before any block headers, is a special block called the start block. Any text will be displayed when the game starts. Any [actions](#actions) will run immediately.

## Actions

```
{set health 100}
```

Actions allow you to change game state, trigger effects, and style text. An action starts with an opening `{`, is followed by the name of the action, may optionally include one or more inputs for the action, then ends with a closing `}`. They trigger whenever the block they are included in triggers.

```
= room is dark =
>> light >>
{set room is not dark}
The light is on. You can see!
```

In addition to writing action inputs between the curly braces (`{}`), display text can be used as an input with an "enclosing" action.

Write the initial action including its opening `{` and closing `}`, then write the display text, and finish with a closing tag that starts with `{/`, repeats the name of the action, and ends with another `}`. Much like HTML, this is commonly used to style text.

```
> insult >
>> brigand >>
{b}What did you say!?{/b}
```

Finally, actions can be used to produce text to display. The names of [variables](#variables) or other expressions can be put between curly braces and whatever text they produce will be displayed.

```
> look >
>> mirror >>
You see {name of player} looking back at you. Not bad.
```

## Variables

Variables are named values that can change over the course of a game. They can be used in actions and state block headers. Their value is typically changed with the "set" action.

```
> shoot >
{set score 3}
You shoot from downtown. It goes in! {score} points!

= score is 3 =
You won!
```

### Creating Variables

Variables are created automatically when used. There is no explicit creation or initialization step. However, there are two rules to keep in mind:

1. While the value of a variable may change, the [type](#variable_types) of value may not. A particular variable must always refer to a value of the same type (i.e. always text or always a number).
2. Each variable must be consumed at least once and set at least once. Variables which are only referenced but never set, or only set but never referenced, are not allowed.

A variable which is referenced before it is explicitly assigned a value will be in a default "unset" state. This exact default value depends on the type (for example, 0 for numbers). If a different starting value is required, set it in a [start block](#start_block).

### Variable Names

Variable names must be unique, may not match the names of any actions or keywords (e.g. not "set", "is", "of", etc), may not _start_ with a number but may include numbers _after_ the first character, and should generally consist of only alphanumeric characters and underscores (\_).

The names are not case sensitive, so "score", "Score", and "SCORE" all refer to the same variable. Spaces are not allowed in variable names, so names with multiple words should use underscores (for example, "opponent_score").

For names which do not use the English alphabet, Unicode characters like 世 or 😀 are supported, but variable names must _not_ include any special reserved characters. Reserved characters include spaces, all special characters printed on a standard QWERTY keyboard other than underscores (~, ., <, +, etc), and any quotes, smart quotes, or apostrophes (", ', 〝, «, etc).

### Variable Types

#### Flags

The most basic variable type is a flag. These are either set or unset and store no other value.

```
> key >
{set unlocked}
You unlocked it!

== unlocked ==
You already unlocked it!
```

If necessary, flag values can be written out explicitly using certain keywords.

Set:

- `true`
- `on`
- `yes`

Unset:

- `false`
- `off`
- `no`

You can also use the "unset" action to revert a variable (flag or otherwise) to its default state.

```
> key >

== locked ==
{unset locked}
You unlocked it!

== not locked ==
{set locked}
You locked it!
```

#### Numbers

Variables can store both whole and decimal numbers. You can use a period as a decimal point, a minus sign to designate a negative number, and either commas or underscores as a thousands separator (if desired).

```
{set count 4}
{set cost 3.50}
{set balance -1000}
{set big_number 1,000,000}
{set bigger_number 1_000_000_000}
```

#### Text

Variables can also hold text. When written within an action or header, text values must have quotation marks on either side.

```
{set name "Alice"}
```

Although straight double quotation marks are recommended, single quotes, backticks, curly quotes, smart quotes, and non-English quotes are all supported as long as the ending quotation mark matches the starting quotation mark.

```
{set city 'Pittsburgh'}
{set state `Pennsylvania`}
{set country “United States of America”}
{set planet «Earth»}
```

You can also use an enclosing action to set text without using quotation marks.

```
{set description}A dark foreboding tower{/set}
```

## Aliases

Aliases are similar to variables, but rather than holding values, they hold a list of possible player inputs. If any part of the text a player inputs is found in the list, input blocks containing that alias are triggered. Aliases are set using the "alias" action, followed by the name of the alias, followed by a text list of possible inputs separated by spaces.

```
{alias greet "greet wave nod hello hi howdy"}
{alias insult}insult yell spurn disparage cast aspersions{/alias}
```

Multi-word inputs can be grouped using quotation marks.

```
{alias greet "greet wave nod 'say hello' 'say hi' 'say howdy'"}
{alias insult}insult yell spurn disparage "cast aspersions"{/alias}
```

Alias declarations respect the conditions of the block they are set in. If set in a start block, the inputs added to an alias will always apply, but inputs added in a state block will only apply when the conditions of that block are met.

```
{alias bounce}bounce jump spring{/alias}

= player in bounce_house =
{alias bounce}step move twitch{/alias}
```

## Objects

Objects are collections of values which typically represent something physical in the game: a room, a sword, the player themselves. The values they contain can be accessed using the "of" keyword or using a colon (":").

```
{set name of door}Iron Door{/set}

> key >
{set door:locked}
The door is locked!

== door:locked ==
You already locked the door!
```

For flag values in an object, the "is" keyword can be used to establish whether or not they are set.

```
> key >

== door is locked ==
{set door is not locked}
You unlocked it!

== door is not locked ==
{set door is locked}
You locked it!
```

### The Player Object

Each game automatically has an object named "player" which represents the main character of the game. This is a useful place to store values like health, skills, and other details specific to the player.

```
{set health of player 100}
```

The [location](#object-location) of the player object is often important for determining the outcome of different conditions.

### Object Location

All objects have a special "location" value which can be set to any other object. This location helps give some underlying structure to the game. It is used to evaluate the "in", "has", and "with" keywords, as well as to determine what the player is and is not able to interact with.

```
{set location of player entrance}
{set door:location entrance}

> open >
== player with door ==
You open the door
```

The "place" action can be used as a shorthand for setting an object's location value.

```
{place player entrance}
{place door entrance}
```

### Objects as Aliases

Objects and aliases may share a name. This is useful to make the game aware of when the player is interacting with an object.

```
{alias door}door entryway hatch{/alias}

> lock >
>> door >>
{set door is locked}
You lock the door!
```

### Objects as State

A non-player object may be used in a state header with no other qualifiers. This is a shorthand for checking that the player is located in that object.

```
> leave >

== cell ==
You can't leave, the cell door is locked!

== dining_room ==
You get up and walk out. You've lost your appetite.
```

### Object Name

Each object has a special "name" value, which contains the text that is displayed when an object is used in an action.

```
{set name of door}The Ancient Iron Door{/set}

> knock >
>> door >>
Meekly, you wrap your knuckles on {door}... no answer.
```

The "name" action can be used as a shorthand for setting an object's name value.

```
{name door}The Ancient Iron Door{/name}
```

### Display Attributes for Objects

Objects have a list of special values which may be used to generate labels, tooltips, and other representations of the object during the game.

- "description"
- "link"
- "color"
- "image"
- "image_icon"
- "image_hero"
- "image_background"
- "sound"
- "sound_background"

The specifics of how these values change the representation of an object will depend on the game engine, but the text they store can be referenced like any other text value.

## Arithmetic

Tale Maker supports basic arithmetic using common math operators including addition (`+`), subtraction (`-`), multiplication (`*`), division (`/`), remainder (`%`), greater than (`>`), less than (`<`), greater than or equal to (`>=`), and less than or equal to (`<=`). All math follows the typical order of operations, with equations in parentheses going first, followed by multiplication, division, and remainder, and then finally addition and subtraction. These operations may only be used with number values and variables that store number values.

```
> shoot >
== score < goal ==
{set score score + 1}
Almost there! Just {goal - score} more points!

== score >= goal ==
You did it! {(score * 3 / 3 - 0) % 1} points!
```

## Comments

Comments are sections of text which are ignored and will not be displayed. They are useful for leaving notes for the people _writing_ the tale, whether yourself or a collaborator. A comment starts with `{!` and continues until a closing `!}`.

```
> search >
You don't find {item {!rename this!}}
{! too short? !}

== cave ==
{!
    TODO: Fill this in.
    This is the big opening to the adventure!
    Make it good!
!}
```

The one place you cannot place a comment is in the middle of quoted text. All characters in quoted text are printed exactly as written. The following action sets the value of "message" to the text _"This is {!not!} a comment"_:

```
{set message "This is {!not!} a comment"}
```

## Escape Characters

When writing text blocks, it is sometimes ambiguous whether a character is meant to be displayed for the player or if it is a part of the Tale Maker syntax. In those cases, a backslash can be used before the character to specify that it is meant to be displayed.

Useful escape characters include:

- `\{` - to display "{"
- `\>` - to display ">" (only needed at the start of a line)
- `\=` - to display "=" (only needed at the start of a line)
- `\\` - to display a backslash
- `\s`, `\n`, `\r`, `\t`, `\v`, `\f` - to display a space, new line, carriage return, tab, vertical tab, or form feed. Since text blocks display most white space exactly as written, these escapes are not typically necessary. One exception is if you want to display empty lines at the start or end of a block.
- Finally, a `\` followed immediately by a line break will _prevent_ the line break from being displayed, causing two lines to display as one.


```
> do_math >
\s


    ...finally \{it} comes to you, is 3 < 2? No!

3
\>
2
\=
\{\{TRUE}}

What a \
Triumph!\n\n\n
```

By contrast, within _quoted_ text there are no escape characters. Every character in quoted text is included exactly as it is typed, including backslashes and line breaks. This means if you want to include a quotation mark in quoted text, you must use a different kind of quotation mark at the start start and end, or else use an enclosing action instead.

```
{set sarcastic_message 'What a "good" job...'}
{set sarcastic_retort}...what a "witty" insult{/set}
```

## Keywords

These keywords can be used to create more complex expressions in block headers and actions.

### :

A colon can be used to reference a value of an object.

```
= player:score is 3 =
Three points!
```

### and

Combines two conditions, specifying that both must be valid.

```
= room is dark and light is not lit =
You cannot see!
```

### has

Determines if the object on the left is the location for the object on the right.

```
> open >
>> door >>
=== player has key ===
You pull out your key, unlock the door, and swing it open!
```

### in

Determines if the object on the right is the location for the object on the left.

```
> press >
>> button >>
=== button in room ===
You press the button and brace yourself for anything...
```

### is

This keyword can be used in different ways depending on the context. If used with non-object variables and values, it determines equality.

```
= score is 3 =
Three points!
```

If used with objects, it determines that a value of an object (usually a flag) is set.

```
= door is locked =
Drat! Have to find another way in.
```

### it

A shorthand specifying a non-player object in the block header. Can only be used if there is exactly one non-player object in the immediately wrapping block header.

```
> leave >
== player in cell ==
You look at the door of {it}. You aren't getting out that way.

=== it is open ===
You peak out the cracked door of {it}. The coast is clear!
```

### not

Negates a condition.

```
= door is not locked =
Sweet! You let yourself in.
```

### of

A way to reference a value of an object. Unlike a colon, the name of the value comes before the name of the object.

```
= score of player is 3 =
Three points!
```

### or

Combines two conditions, specifying that either may be valid.

```
> break >
>> door >>
=== player is strong or player is angry ===
You smash down the door!
```

### with

Determines if two objects have the same location.

```
> pull >
>> lever >>
=== lever with player ===
You give the lever a pull!
```

## Available Actions

This is a list of actions built into Tale Maker. Game engines may specify additional actions which trigger visual effects, further styling, or other effects specific to the engine.

### alias

Adds a space separated list of possible input text to a named alias.

```
{alias greet "greet wave hail"}
{alias hit}strike punch smash{/alias}
```

Use matching quotation marks within the list to specify input text that should include spaces.

```
{alias iron_door "'iron door' 'big door' 'entrance door'"}
{alias red_key}"red key" "ruby key"{/alias}
```

### b

Styles enclosed text as bold.

```
Here comes my {b}MEGA{/b} move!
```

### chain

Enclose around multiple "choice" actions to display text from each in series. The first "choice" without a condition or with a valid condition is displayed. Afterwards, if the same "chain" action is run again, it will ignore the previously made choice and instead select the next one.

```
{chain}
{choice player is strong}Hey, big fella{/choice}
{choice}Hey{/choice}
{choice}Howdy{/choice}
{choice}How are ya?{/choice}
{/chain}
```

If all valid choices have already been displayed, the chain will repeat. All text within a "chain" must be wrapped in a "choice".

### chance

Similar to [chain](#chain), but unordered. The "chance" action randomly selects a "choice" to display from among all valid choices which have not yet been displayed. Once all valid choices have been displayed, the list will repeat. All text within a "chance" must be wrapped in a "choice".

### choice

Encloses text which may be displayed by a [choose](#choose), [chain](#chain), or [chance](#chance) action. Only one "choice" will display depending on the rules of the enclosing action. Optionally may include a condition. If included, the choice will not display unless the condition is valid.

### choose

Enclose around multiple "choice" actions to display only text from the first with a valid condition. All later "choice" actions are ignored. May contain a final "choice" without a condition which will act as a default if no earlier "choice" is valid.

```
{choose}
{choice player is strong}You bust through!{/choice}
{choice player is clever}You solve the puzzle!{/choice}
{choice player is fast}You run, no one can catch you!{/choice}
{choice}You're screwed{/choice}
{/choose}
```

Unlike [chain](#chain) and [chance](#chance), "choose" has no memory of which choices have already been displayed and will always select the first valid one. All text within a "choose" must be wrapped in a "choice".

### do

Triggers another block, including triggering any contained actions and displaying any text. Accepts aliases as an input, and will determine which block to trigger using those aliases and the current game state.

```
> run >
You set off running as fast as you can, the sounds of pursuit fresh on your heels.

== player is fast ==
{set escaped}
You set off running as fast as you can, the sounds of pursuit fading into the distance.

> jump >
You burst through the window, the tinkling of glass mixing with shouts of confusion from your captors.
{set player is fast}
{do run}
```

### i

Styles enclosed text as italic.

```
And what is {i}that{/i} supposed to mean??
```

### if

An action which takes a condition and text. The text will only be displayed if the condition is valid.

```
> greet >
"{if player is intimidating}{i}*gulp*{/i}{/if} Hello there stranger," squeaks the little goblin.
```

### name

A shorthand for setting the "name" value of an object.

```
{name ship}Excelsior{/name}
```

### place

A shorthand for setting the "location" value of an object

```
{place lever entrance}
```

### set

Sets the value of a variable or an object value. For flag variables, no value needs to be specified.

```
{set having_fun}
{set score 3}
{set description of door}A big heavy door{/set}
```

### title

Styles enclosed text as a title or heading.

```
{title}An Unexpected Development{/title}
```

### unset

Reverts a variable or object value back to an unset state. For numbers this means a value of 0, and for text this means no text. Primarily meant to be used with flag variables.

```
{unset having_fun}
{unset door:locked}
```

## Built-in Variables

### any

A special alias which matches every possible player input that does not match another player input in the same wrapping block. Useful for providing fallback text.

```
> any >
I don't know what you mean by that.
```

### player

A special object representing the player. Particularly important for determining game location.

```
> lever >
== player not with it ==
You can't throw the lever, it's not here!
```

### repeat

A special variable that is set if the immediately wrapping block has been triggered at least once already.

```
> greet >
You say hello and introduce yourself.

== repeat ==
You give a quick wave.
```

### tale

A special object representing the overall game itself. Used mostly to specify information about the game, like a name that can be displayed on a list for the player to choose from.

```
{name tale}My Awesome Adventure{/name}
```
