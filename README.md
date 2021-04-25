# piston_bot

A Telegram bot that will run code for you. Made using [piston][1].

Available as [@iruncode_bot](https://t.me/iruncode_bot) on telegram.

## Examples

### Basic example

- Input:

  ```python
  /run python3
  print('Hi')
  for i in range(5):
      print(i)
  ```

- Response:

  **Language:**

  ```console
  python3
  ```

  **Code:**

  ```console
  print('Hi')
  for i in range(10):
      print(i)
  ```

  **Output:**

  ```console
  Hi
  0
  1
  2
  3
  4
  ```

### With user input

- Input:

  ```python
  /run py
  print(input())
  /stdin
  Hello
  ```

- Response:

  **Language:**

  ```console
  py
  ```

  **Code:**

  ```console
  print(input())
  ```

  **Output:**

  ```console
  Hello
  ```

## Deploy your own

You'll need [go](https://golang.org) installed.

- Create a telegram bot, and copy its token.
- Run the following in your terminal:

  ```bash
  go build ./cmd/bot

  export TOKEN=<your telegram bot token>
  ./bot         # starts the bot
  ```

[1]: https://github.com/engineer-man/piston
