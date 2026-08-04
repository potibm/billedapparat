# Billedapparat: Styling

The Beamer can be customized for your event. This can be done by overwriting css styles.

Mount a style directory into the `billedapparat` docker container at `/app/data/style`. Most probably you already have `/app/data` already mounted somewhere.

Create a file called `custom.css` within this directory, which will be used in the Frontend. We are using BEM (Block Element Modifier) naming convention.

## Example

````css
.slide-news {
    background: url("/style/assets/background.png") no-repeat center center fixed !important;
}

.slide-news__content {
    font-family:'Comic Sans MS', cursive, sans-serif !important;
}

.slide-news__content h1 {
    color: #00ff00 !important;
}

.formatted-text__hashtag {
    color: #ff00ff !important;
}
````

## Assets

You may add and reference further assets to the style directory, as you can see from the `assets/background.png` above. 
