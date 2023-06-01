# Muzikas - A Discord Music Bot

Muzikas is a simple, and easy-to-use Discord bot that brings music to your server. It is written in Go, offering a responsive and resource-friendly music bot.

## Features

- Play music directly from YouTube.
- Queue management commands (play, queue, pause, skip, etc.).

## Prerequisites

Before you begin, ensure you have met the following requirements:

- You have installed the latest version of Go.
- You have a Discord bot token. 

## Setting up Muzikas

To install Muzikas, follow these steps:

1. Clone the repository:
    ```shell
    git clone https://github.com/yourusername/Muzikas.git
    cd Muzikas
    ```

2. Set your Discord bot token as an environment variable:
    ```shell
    export TOKEN=your-discord-bot-token
    ```

3. Build the project:
    ```shell
    go build -o Muzikas cmd/main.go
    ```

4. Run the bot:
    ```shell
    ./Muzikas
    ```

## Using Muzikas

To use Muzikas, add the bot to your server and use the following commands:

- `!play <youtube-url>`: Plays the specified YouTube video's audio.
- `!queue <youtube-url>`: Adds the Youtube video's audio to queue.
- `!pause`: Pauses the current audio.
- `!unpause`: Resumes the paused audio.
- `!stop`: Stops the playback and disconnects the bot voice connection
- `!skip`: Skips the current audio and plays the next one in the queue.

## Contact

If you want to contact me, you can reach me at `fabijan.zulj@gmail.com`.

