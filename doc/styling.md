# Billedapparat: Styling

The Beamer can be customized for your event. This can be done by overwriting css styles.

Mount a style directory into the `billedapparat` docker container at `/app/data/style`. Most probably you already have `/app/data` already mounted somewhere.

Create a file called `custom.css` within this directory, which will be used in the Frontend. We are using BEM (Block Element Modifier) naming convention.

## Example

```css
.slide-news {
  background: url("/style/assets/background.png") no-repeat center center fixed !important;
}

.slide-news__content,
.slide-urgent__content {
  font-family: "Comic Sans MS", cursive, sans-serif !important;
}

.slide-news__title {
  color: #00ff00 !important;
}

.slide-urgent__title {
  color: #ff0000 !important;
}

.formatted-text__hashtag {
  color: #ff00ff !important;
}
```

## Assets

You may add and reference further assets to the style directory, as you can see from the `assets/background.png` above.

## Available BEM Classes

Each slide type follows the BEM naming convention: `.slide-{type}__{element}`. Use these hooks to override styles for specific elements.

### News (`slide-news`)

| Class                    | Element                              |
| ------------------------ | ------------------------------------ |
| `.slide-news`            | Root container                       |
| `.slide-news__container` | Wrapper for content layout           |
| `.slide-news__content`   | Content area (title + markdown body) |
| `.slide-news__title`     | The `<h1>` headline                  |
| `.slide-news__body`      | Markdown body content                |
| `.slide-news__fallback`  | Shown when no body content exists    |

### Urgent News (`slide-urgent`)

| Class                      | Element                              |
| -------------------------- | ------------------------------------ |
| `.slide-urgent`            | Root container                       |
| `.slide-urgent__container` | Wrapper for content layout           |
| `.slide-urgent__content`   | Content area (title + markdown body) |
| `.slide-urgent__title`     | The `<h1>` headline                  |
| `.slide-urgent__body`      | Markdown body content                |
| `.slide-urgent__fallback`  | Shown when no body content exists    |

### Timetable (`slide-timetable`)

| Class                               | Element                              |
| ----------------------------------- | ------------------------------------ |
| `.slide-timetable`                  | Root container                       |
| `.slide-timetable__container`       | Wrapper for content layout           |
| `.slide-timetable__content`         | Content area (title + markdown body) |
| `.slide-timetable__title`           | The `<h1>` headline                  |
| `.slide-timetable__table`           | The whole table                      |
| `.slide-timetable__cell-start-time` | Cells with start time content        |
| `.slide-timetable__cell-end-time`   | Cells with end time content          |
| `.slide-timetable__cell-category`   | Cells with category                  |
| `.slide-timetable__cell-title`      | Cells with title                     |
| `.slide-timetable__cell-location`   | Cells with location                  |
| `.slide-timetable__fallback`        | Shown when no body content exists    |

### Sponsor (`slide-sponsor`)

| Class                      | Element                           |
| -------------------------- | --------------------------------- |
| `.slide-sponsor`           | Root container                    |
| `.slide-sponsor__image`    | The sponsor image (`<img>`)       |
| `.slide-sponsor__fallback` | Shown when no image is configured |

### Social (`slide-social`)

| Class                     | Element                                 |
| ------------------------- | --------------------------------------- |
| `.slide-social`           | Root container                          |
| `.slide-social__media`    | Wrapper around the media image          |
| `.slide-social__image`    | The media image (`<img>`)               |
| `.slide-social__header`   | Author header (avatar, name, timestamp) |
| `.slide-social__body`     | The post text content                   |
| `.slide-social__fallback` | Shown when no body content exists       |
