# discord-links-bot

A discord bot for replacing/trimming common links with more appropriate versions.

Message embeds are hidden when possible.

## Supported links
- twitter links are replaced with fxtwitter links (bot hides original message's embed)
- media.discordapp.net links are replaced with cdn.discordapp.com (bot does nothing to the original message, for now)
- amazon product links are trimmed down to the bare minimum required to visit the product page (hides link embed)
- reddit links are shortened to the bare minimum required to visit the submission (hides link embed)
- instagram links are embedded using ddinstagram, removing any link trackers (only enabled for reels)

### Examples

#### Twitter

```
https://twitter.com/blablabla/status/12345678910?cheese=doodles
```
Turns into

```
https://fxtwitter.com/blablabla/status/12345678910
```

#### Discord

```
https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4
```

Turns into

```
https://cdn.discordapp.com/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4
```

#### Amazon

```
https://www.amazon.co.uk/dp/B09B1HMJ9Z/ref=as_li_ss_tl?ie=UTF8&smid=ATVPDKIKX0DER&th=1&linkCode=sl1&tag=sec2002-20
https://www.amazon.com/dp/B09B1HMJ9Z/ref=as_li_ss_tl?ie=UTF8&smid=ATVPDKIKX0DER&th=1&linkCode=sl1&tag=sec2002-20
```

Turns into
```
https://www.amazon.co.uk/dp/B09B1HMJ9Z
https://www.amazon.com/dp/B09B1HMJ9Z
```

#### Reddit

```
https://www.reddit.com/r/truetf2/comments/107nizk/what_makes_a_lot_of_the_configuration/?utm_medium=ios_app
```
Turns into
```
https://www.reddit.com/r/truetf2/comments/107nizk/what_makes_a_lot_of_the_configuration/
```

#### Instagram

```
https://www.instagram.com/reel/CztdYC8ryw7/?igshid=abcdefghujkl==
```
Turns into
```
https://www.ddinstagram.com/reel/CztdYC8ryw7
```
