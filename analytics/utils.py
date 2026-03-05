import matplotlib.ticker as ticker


def label(scale: int, ef: int) -> str:
    return f"S{scale}/EF{ef}"


def nodes(scale: int) -> int:
    return 2**scale


def sci(ax, axis="y"):
    fmt = ticker.FuncFormatter(lambda x, _: f"{x:.2e}" if x != 0 else "0")
    if axis == "y":
        ax.yaxis.set_major_formatter(fmt)
    else:
        ax.xaxis.set_major_formatter(fmt)
