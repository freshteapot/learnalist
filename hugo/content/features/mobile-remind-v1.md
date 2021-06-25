---
title: Add, Remind, Repeat
type: features
css_include: ["main"]
---

# Short version
A little app, trying to help YOU remember things

# Status
## iOS
- available via TestFlight, with bugs
- Send me a message to get access

## Android
- Not yet available (but soon), in the meantime you can always use the website
- Send me a message if your interested to test it

# Bugs
- [#229](https://github.com/freshteapot/learnalist-api/issues/229) You need to accept "push notifications" or things dont work.
# Features
- Add an entry
- Remove an entry
- Review entries
- Get reminders to your phone to aid you in your learning
- Uses spaced repetition delivered via https://learnalist.net
- Set a daily reminder to motivate you or pat you on the back for adding more entries
# But first, what is spaced repetition
People have concluded that being reminded of something over time helps remember it,

> taking it from your short term memory and moving it to your long term memory.

Honestly, [the overview on wikipedia](https://en.wikipedia.org/wiki/Spaced_repetition) is the best place to dive into the what.

Searching the internet for "spaced repetition" will show you research papers, blog posts and more that will help answer your questions.

At learnalist.net, with your help, we will remind you of something sooner or later based on time intervals.

We use the following time intervals

| Level | Over time |
| --- |---|
| 0 | 1 hr |
| 1 | 3 hr |
| 2 | 8 hr |
| 3 | 1 day |
| 4 | 3 days |
| 5 | 7 days |
| 6 | 14 days |
| 7 | 30 days |
| 8 | 60 days |
| 9 | 120 days |

### just remember
**You add an entry and we will handle the time intervals**.

# Using the app

## Add entry
- Login
- Click +
- Add entry
- Wait 1 hour and we will notify you to review your entry

Keep reading to learn what "review your entry" means.

### Examples of an entry
#### Learning a language
I am learning Norwegian and I want to learn chocolate in English means sjokolade in Norwegian.

- Open the app
- Click +
- Write "sjokolade"
- Tick "Add meaning / defenition"
- Write "chocolate"
- Click Add
- In 1 hour we will notify you about sjokolade

#### Expanding your vocabulary
I just discovered the word ridiculous, its fun to say, and I want to use it in future sentences

- Open the app
- Click +
- Write "ridiculous"
- Tick "Add meaning / definition"
- Write "Astonishing; unbelievable"
- Click Add
- In 1 hour we will notify you about ridiculous


You can add entries via the web, browser extension or the app

## Review entry
- A notification has arrived
- You open the app
- Click on 🧠 + 💪
- Read the entry
- Be honest to yourself, did you remember it? (its okay if you didnt, it is all about you)
- Tap on it to show the meaning or definition
- If you remembered it instantly, click "later" (you can also click sooner if you want)
- If you took sometime to remember it, click "sooner"
- If you didnt remember it, click "sooner"

### Later
When you click later, you are helping us know to increase the time until we show it again

### Sooner
When you click sooner, you are helping us know to decrease the time until we show it again

## Remove entry
- Open the app
- Tap on an entry
- Click the red bin / trash can and it will be forever deleted

# Reference
- [Spaced Repetition on wikipedia](https://en.wikipedia.org/wiki/Spaced_repetition)
- [ridiculous on wiktionary](https://en.wiktionary.org/wiki/ridiculous)

# Give us feedback
We would love to hear from you, [please create an issue on github](https://github.com/learnalist/support/issues/new).
