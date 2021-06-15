<script>
    import { scaleLinear } from "d3-scale";
    export let data;
    export let xTicks;
    export let yTicks;
    export let accessor;
    let width = 2000;
    let height = 350;
    let padding = {
        left: 10,
        bottom: 10,
        top: 10,
        right: 10,
    };

    $: xScale = scaleLinear()
        .domain([0, xTicks.length])
        .range([padding.left, 2000 - padding.right]);

    $: yScale = scaleLinear()
        .domain([0, Math.max(...yTicks)])
        .range([height - padding.top, padding.bottom]);
</script>

<div class="container">
    <svg>
        <g transform="translate({padding.left}, {height - padding.bottom})">
            {#each xTicks as tick, i}
                <g transform="translate({xScale(i)}, 2)">
                    <text transform="rotate(90)">{tick}</text>
                </g>
            {/each}
        </g>
        <g transform="translate(0, 0)">
            {#each yTicks as tick, i}
                <g transform="translate(0, {yScale(tick)})">
                    <line x1="20" x2="100%" />
                    <text y="30">{tick}</text>
                </g>
            {/each}
        </g>
        <g class="bars" transform="translate({padding.left})">
            {#each data as d, i}
                <rect
                    fill={"#fe6565"}
                    x={xScale(i)}
                    y={yScale(accessor(d))}
                    width={10}
                    height={height - padding.bottom - yScale(accessor(d))}
                />
            {/each}
        </g>
    </svg>
</div>

<style>
    .container {
        width: 100vw;
        overflow-x: auto;
    }

    svg {
        width: 2000px;
        height: 400px;
    }

    text {
        font-size: 12px;
    }
    line {
        stroke: #aaa;
        stroke-dasharray: 2;
    }
</style>
