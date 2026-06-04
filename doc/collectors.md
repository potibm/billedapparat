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

## Twitch

To retrieve the avatar pictures, we will need a Twitch Client ID and Client Secret. If you do not need this: simply skip this step and leave the config values empty.

### Prerequisites
Twitch requires **Two-Factor Authentication (2FA)** to be enabled on your account before you can access the developer tools. 
* If you haven't set up 2FA yet, go to your [Twitch Security Settings](https://www.twitch.tv/settings/security) and enable it.

### 1. Log into the Developer Console
1. Open your web browser and go to the [Twitch Developer Console](https://dev.twitch.tv/console).
2. Log in with your standard Twitch account credentials.
3. If this is your first time, you may be asked to accept the Twitch Developer Agreement.

### 2. Register a New Application
1. Click on the **Applications** tab at the top of the screen.
2. Click the purple **+ Register Your Application** button on the right side.

### 3. Fill out the Application Details
You will see a form. Fill it out exactly like this:

* **Name**: Choose a unique name for your application (e.g., `MyPersonalChatBot123`). 
  * *Note: Twitch requires this name to be globally unique. You cannot use the word "Twitch" in the name.*
* **OAuth Redirect URLs**: Type `http://localhost` into the field and click the **Add** button. 
* **Category**: Select **Application Integration** (or **Chat Bot**) from the dropdown menu.
* **Client Type**: Select **Confidential**.

Once you have filled out all fields, click the **Create** button at the bottom.

### 4. Get your Credentials
1. You are now back on the Applications list. Find the app you just created and click the **Manage** button next to it.
2. **Client ID**: Scroll down to the "Client ID" section. Copy this long string of letters and numbers and paste it into our app's settings.
3. **Client Secret**: Click the **New Secret** button. A popup will appear asking if you are sure; click **OK**. 
4. Your `Client Secret` will now be displayed. **Copy it immediately** and paste it into our app's settings.

> ⚠️ **IMPORTANT:** The Client Secret is like a password. Twitch will only show it to you **once**. If you close the page without copying it, you will have to click "New Secret" again to generate a new one.

````yaml
collectors:
    twitch:
        api_key: 123***redacted**xyz
        enabled: true
        type: slide
        channel: "endless_demoshow"
        client_id: ""
        client_secret: ""
````