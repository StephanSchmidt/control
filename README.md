# control

Generate SVG diagrams from plain text. Define boxes on a grid, connect them with arrows, and get clean vector output.

Written for the diagrams in my [Theory of Control of Software Engineering](https://www.tabulamag.com/p/introduction-to-theory-of-control), designed to be embedded into [sli.dev](https://sli.dev) markdown presentations.

### AI-friendly by design

`control` is built to work well with AI coding agents. The plain-text input format is easy for LLMs to generate and modify — no GUI, no binary formats, no complex APIs. The CLI provides comprehensive `--help` output with usage examples, and the `--debug` flag outputs structured JSON for programmatic inspection. Feed an AI agent the format spec, and it can produce and iterate on diagrams autonomously.

![Scrum workflow example](examples/scrum.svg)

## Tutorial

See the [step-by-step tutorial](tutorial/tutorial.md) for a progressive guide covering all features — from basic boxes to full diagrams with custom colors, groups, containers, and arrow flow.

## Install

### Homebrew (macOS/Linux)

```
brew tap StephanSchmidt/control
brew install control
```

### Go

```
go install control@latest
```

### Build from source

```
make install
```

## Quick start

Create a text file (`diagram.txt`):

```
1,1: Planning
>+2,+1: Development
>+2,0: Review
>+2,-1: Ship
```

Generate an SVG:

```
control --diagram diagram.txt --out diagram.svg
```

## Diagram format

A diagram file contains box definitions and optional arrow definitions. Lines starting with `#` are comments.

### Box syntax

```
[id:] x,y[,width[,height]]: Label[, style] [@Group]
```

| Field    | Description                                              |
|----------|----------------------------------------------------------|
| `id`     | Optional identifier for arrows (alphanumeric, `_`, `-`)  |
| `x,y`    | Grid coordinates (starting from 1)                       |
| `width`  | Grid width (default: 2, supports fractions like `1/2`)   |
| `height` | Grid height (default: 1)                                 |
| `Label`  | Display text inside the box                              |
| `style`  | Optional style codes (see below)                         |
| `@Group` | Optional group assignment (see below)                    |

### Relative coordinates

Use `+N` or `-N` for positions relative to the previous box. Use `0` as shorthand for `+0`.

```
1,1: Box A
+2,+1: Box B      # resolves to 3,2
0,+1: Box C        # resolves to 3,3
```

### Auto-arrows (`>` prefix)

Prefix coordinates with `>` to create an arrow from the previous box:

```
1,1: Start
>+1,+1: Next       # creates arrow Start -> Next
```

### Touch-left (`|` prefix)

Prefix with `|` to visually attach a box to the previous one (no gap):

```
1,1: First
|+2,0: Second      # Second touches First on its left side
```

### Styles

| Code  | Effect                          |
|-------|---------------------------------|
| `g`   | Gray background                 |
| `p`   | Purple background               |
| `lp`  | Light purple background         |
| `rb`  | Red border (3px)                |
| `rt`  | Red text                        |
| `nbb` | No background, no border        |
| `2t`  | Double text size (48px)         |

Combine styles with dashes: `rb-g`, `nbb-rt-2t`

### Groups

Visually group boxes with a surrounding dashed rectangle:

```
1,2: Alice @Team
>+2,0: Bob, p @Team
@Team: Engineering
```

Append `@GroupName` to assign a box to a group. Define `@GroupName: Label` to set a custom label (defaults to the group name).

### Containers

Group elements with relative coordinates using `[...]` brackets. Coordinates inside a container are relative to the container's position. Moving the container moves all its elements.

```
G: 3,2 [
    X: 0,0: Alpha
    +2,0: Beta
    Y: 0,2: Gamma
    X -> Y
]
```

- `0,0` resolves to the container origin (3,2)
- `+1,0` is relative to the previous box (works as usual)
- `4,4` is offset from the container origin (resolves to 7,6)
- Box IDs inside containers are scoped: `X` becomes `G.X`
- Arrows inside containers use local IDs (`X -> Y` auto-scopes to `G.X -> G.Y`)
- Outside, reference container boxes with `ContainerID.BoxID` (e.g., `A -> G.X`)
- Containers are purely organizational — no visual border is drawn (use `@Group` for that)
- Nesting is not supported

### Arrows

```
from_id -> to_id
```

Arrow lines can appear anywhere in the file. Arrows route automatically using orthogonal segments.

### Frontmatter

Optional metadata at the top of the file, enclosed between `---` delimiters:

```
---
font: fonts/MyFont.woff2
x-label: Time
y-label: Control
legend: p = In Progress
legend: g = Completed
color: green = #00FF00
---
```

| Key       | Description                                |
|-----------|--------------------------------------------|
| `font`    | Custom font file (WOFF2 format)            |
| `x-label` | X-axis label (omit to hide axis)          |
| `y-label` | Y-axis label (omit to hide axis)          |
| `legend`  | Legend entry: `style = description`        |
| `color`   | Custom color: `name = #hex`                |

Custom colors can be used as style codes (`green` for background, `greent` for text color).

## CLI options

```
--diagram <file>    Input diagram file (default: examples/diagram.txt)
--out <file>        Output SVG file (default: examples/diagram.svg)
--stretch <float>   Horizontal stretch factor (default: 1.0)
--vertical-gap <f>  Vertical gap in grid units (default: 0.5)
--font <file>       Custom font file (WOFF2) to embed in SVG
--debug <file>      Output debug information to JSON file
```

## Example

The Scrum workflow diagram (`examples/scrum.txt`):

```
1,1: OKRs
>+1,+1: Business Initiatives
>+1,+1: Backlog Planning
>+1,+1: Sprint Planning
>+1,+1,1: Daily
>+1,+1: Developing, p
>+2,-1,1/2:
>0,+1: Developing, p
|+2,0,1: Idle, lp
>+1,-1,2: Sprint Review
9,1: SCRUM, nbb-rt-2t
```

Generate it:

```
control --stretch 0.8 --diagram examples/scrum.txt --out examples/scrum.svg
```

## Development

```
make build          # Build binary to bin/
make install        # Install to $GOPATH/bin
make test           # Run tests
make lint           # Run vet + staticcheck
make example        # Build and render the scrum example
```

## License

MIT
