# Billedapparat: Collectors

## Mastodon

You will need to create an application unter Preferences > Development.

Give it a name like "Billedapparat", provide a website and redirect url and chose as scope:

- read:notifications
- read:statuse

Copy the "Your access token" after saving.

````
collectors:
    mastodon:
        api_key: 123***redacted**xyz
        enabled: true
        host: mastodon.social
        access_token: 123***redacted**xyz
        tag: demoscene
````

