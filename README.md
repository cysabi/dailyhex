# play wordle, with hex colors ~ right from your terminal!

⏰ resets each day @ 11am rc time <3

🎨 curious how hex codes work? check out https://colorrush.io/learn

🧋 served with https://charm.sh/bubbletea

**never played wordle? here's a quick guide:**

```go
Gray   { wrong val, wrong pos }
Yellow { right val, wrong pos }
Green  { right val, right pos }

/*
this game is works best in a truecolor terminal!
unsure abt your term? the vscode one works
getting a false negative? u can run `ssh -t <hostname> truecolor` to override the auto-detect

any other questions or concerns? search `ssh hex.recurse.cloud` on zulip!
*/
```
