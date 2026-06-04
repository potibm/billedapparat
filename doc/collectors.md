# Billedapparat: Collectors

## Mastodon

You will need to create an application unter Preferences > Development.

Give it a name like "Billedapparat", provide a website and redirect url and chose as scope:

- read:notifications
- read:statuse

Copy the "Your access token" after saving.

````yaml
collectors:
    mastodon:
        api_key: 123***redacted**xyz
        enabled: true
        type: slide
        host: mastodon.social
        access_token: 123***redacted**xyz
        tag: demoscene
````

## Pouet

````yaml
collectors:
    pouet:
        api_key: 123***redacted**xyz
        enabled: true
        type: slide
        poll_interval: 5   # in minutes
        keywords:
            - evoke
            - ovolo
            - zvokz
````

## Bluesky

````yaml
collectors:
    blueysky:
        api_key: 123***redacted**xyz
        enabled: true
        type: slide
        hashtags:
            - #evoke
            - #demoscene
            - #demoparty
````

## Discord

This guide provides step-by-step instructions on how to create a Discord bot in the Developer Portal, configure the necessary permissions (Intents), and invite the bot to a server.

### 1. Create a Bot in the Developer Portal

1. Open the [Discord Developer Portal](https://discord.com/developers/applications) and log in with your Discord account.
2. Click the **"New Application"** button in the top right corner.
3. Enter a suitable name for your application, accept the terms of service, and click **"Create"**.
4. Navigate to the **"Bot"** tab in the left sidebar.
5. *(Optional)* Here you can adjust the bot's username and upload a profile picture.
6. Click **"Reset Token"** to generate the secret bot token. Store it in the discord collector configuration as `bot_token`.    
    **Never** share this token publicly!

### 2. Enable Privileged Gateway Intents

In order for the bot to actually "read" the content of chat messages, a specific security clearance must be enabled within Discord.

1. Stay on the **"Bot"** tab and scroll down to the **"Privileged Gateway Intents"** section.
2. Enable the toggle for the **"Message Content Intent"**.
3. Click the green **"Save Changes"** button at the bottom.

### 3. Generate the Invite Link (OAuth2)

To bring the bot into a server, we need a special invitation URL containing the exact permissions required.

1. Switch to the **"OAuth2"** tab in the left sidebar and select **"URL Generator"** right below it.
2. In the **"Scopes"** box, check the box for `bot`.
3. A second box named **"Bot Permissions"** will appear below.
4. Under *General Permissions*, check the box for **`View Channels`** (which also implies Read Messages). 
5. Scroll all the way down. You will see a **"Generated URL"**. Click the **"Copy"** button next to it.

### 4. Invite the Bot to a Server

Using the generated URL, the bot can now be added to a server.

**Important prerequisite regarding server permissions:**
To invite a bot to a server, the person executing the link must have either the **"Manage Server"** or **"Administrator"** permission on that specific server. 

**Variant A: For your own server**
1. Open a new tab in your web browser and paste the copied URL.
2. Select your own server from the "Add to Server" dropdown menu.
3. Click **"Continue"**, review the permissions, and click **"Authorize"**.

**Variant B: For a server you don't manage**
1. Simply send the copied URL to an Administrator or Moderator of the target server.
2. The administrator clicks the link and selects their server to authorize the bot.

Once the process is complete, the bot will appear in the member list of the respective server. It will come online as soon as your Go backend script is started with the correct token.

### 5. Retrieve the channel ID

Enable the "Developer Mode" in the Settings. 

Now you can right-click on the channel and select "Copy Channel ID" from the popup.

````yaml
collectors:
    discord:
        api_key: 123***redacted**xyz
        enabled: true
        type: slide
        bot_token: bot-token-secret
        channel_id: "1000524269958213747"
````