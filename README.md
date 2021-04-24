# piston_bot

A Telegram bot that will run code for you. Made using [piston][1].

Available as [@iruncode_bot](https://t.me/iruncode_bot) on telegram.

## Example

- Input:

```python
/run py
print('bruh')
for i in range(10):
    print(i)
```

- Response:

**Code:**

```console
print('bruh')
for i in range(10):
    print(i)
```

**Output:**

```console
bruh
0
1
2
3
4
5
6
7
8
9
```

## Deploy your own

You'll need [go](https://golang.org) installed.

- Create a telegram bot, and copy its token.
- Run the following in your terminal:

  ```bash
  go build ./cmd/bot

  export TOKEN=<the bot token>
  ./bot # this runs the bot
  ```

[1]: https://github.com/engineer-man/piston
