import pandas as pd
import matplotlib.pyplot as plt
from pathlib import Path
from utils import sci, label, nodes
import numpy as np


def plot_teps_vs_scale(summ: pd.DataFrame, outdir: Path):
    """Harmonic-mean TEPS vs graph scale, one line per edge factor."""
    fig, ax = plt.subplots(figsize=(8, 5))

    for ef, grp in summ.groupby("edge_factor"):
        grp = grp.sort_values("scale")
        ax.plot(grp["scale"], grp["teps_harmonic_mean"], marker="o", label=f"EF={ef}")

    ax.set_xlabel("Scale (log₂ vertices)")
    ax.set_ylabel("Harmonic-Mean TEPS")
    ax.set_title("Graph500 — TEPS vs Scale")
    ax.legend(title="Edge Factor")
    sci(ax)
    fig.tight_layout()
    fig.savefig(outdir / "teps_vs_scale.png")
    plt.close(fig)
    print("  Saved teps_vs_scale.png")


def plot_teps_box(runs: pd.DataFrame, outdir: Path):
    """Box plot of per-run TEPS for each (scale, edge_factor) configuration."""
    if runs.empty:
        print("  [skip] No per-run data for box plot.")
        return

    configs = runs.assign(
        config=runs.apply(lambda r: label(int(r.scale), int(r.edge_factor)), axis=1)
    )
    order = sorted(configs["config"].unique())

    data = [configs[configs["config"] == c]["teps"].values for c in order]

    fig, ax = plt.subplots(figsize=(max(6, len(order) * 1.4), 5))
    ax.boxplot(
        data,
        labels=order,
        patch_artist=True,
        boxprops=dict(facecolor="#aec6e8", color="#333"),
        medianprops=dict(color="crimson", linewidth=2),
    )
    ax.set_xlabel("Configuration (Scale / Edge Factor)")
    ax.set_ylabel("TEPS per BFS run")
    ax.set_title("TEPS Distribution per Configuration")
    plt.xticks(rotation=30, ha="right")
    sci(ax)
    fig.tight_layout()
    fig.savefig(outdir / "teps_boxplot.png")
    plt.close(fig)
    print("  Saved teps_boxplot.png")


def plot_time_vs_scale(summ: pd.DataFrame, outdir: Path):
    """Median BFS time vs scale with min/max error bars."""
    fig, ax = plt.subplots(figsize=(8, 5))

    for ef, grp in summ.groupby("edge_factor"):
        grp = grp.sort_values("scale")
        yerr = np.array(
            [
                grp["time_median_s"] - grp["time_min_s"],
                grp["time_max_s"] - grp["time_median_s"],
            ]
        )
        ax.errorbar(
            grp["scale"],
            grp["time_median_s"],
            yerr=yerr,
            marker="s",
            capsize=4,
            label=f"EF={ef}",
        )

    ax.set_xlabel("Scale (log₂ vertices)")
    ax.set_ylabel("Median BFS Time (s)")
    ax.set_title("BFS Time vs Scale")
    ax.legend(title="Edge Factor")
    fig.tight_layout()
    fig.savefig(outdir / "time_vs_scale.png")
    plt.close(fig)
    print("  Saved time_vs_scale.png")


def plot_nedge_vs_scale(summ: pd.DataFrame, outdir: Path):
    """Median edges traversed vs number of vertices."""
    fig, ax = plt.subplots(figsize=(8, 5))

    for ef, grp in summ.groupby("edge_factor"):
        grp = grp.sort_values("scale")
        n_vertices = grp["scale"].apply(nodes)
        ax.plot(n_vertices, grp["nedge_median"], marker="^", label=f"EF={ef}")

    ax.set_xlabel("Number of Vertices (2^scale)")
    ax.set_ylabel("Median Edges Traversed")
    ax.set_title("Edges Traversed vs Graph Size")
    ax.set_xscale("log", base=2)
    ax.legend(title="Edge Factor")
    sci(ax)
    fig.tight_layout()
    fig.savefig(outdir / "nedge_vs_scale.png")
    plt.close(fig)
    print("  Saved nedge_vs_scale.png")


def plot_construct_time(summ: pd.DataFrame, outdir: Path):
    """CSR graph construction time vs scale."""
    if "construct_time_s" not in summ.columns:
        print("  [skip] No construct_time_s in data.")
        return

    fig, ax = plt.subplots(figsize=(8, 5))

    for ef, grp in summ.groupby("edge_factor"):
        grp = grp.sort_values("scale")
        ax.plot(grp["scale"], grp["construct_time_s"], marker="D", label=f"EF={ef}")

    ax.set_xlabel("Scale (log₂ vertices)")
    ax.set_ylabel("Construction Time (s)")
    ax.set_title("CSR Graph Construction Time vs Scale")
    ax.legend(title="Edge Factor")
    fig.tight_layout()
    fig.savefig(outdir / "construct_time_vs_scale.png")
    plt.close(fig)
    print("  Saved construct_time_vs_scale.png")


def plot_teps_heatmap(summ: pd.DataFrame, outdir: Path):
    """Heatmap of harmonic-mean TEPS over scale × edge factor grid."""
    pivot = summ.pivot_table(
        index="scale", columns="edge_factor", values="teps_harmonic_mean"
    )
    if pivot.empty:
        return

    fig, ax = plt.subplots(
        figsize=(max(5, len(pivot.columns) * 1.5), max(4, len(pivot) * 0.8))
    )
    im = ax.imshow(pivot.values, aspect="auto", cmap="YlOrRd")
    ax.set_xticks(range(len(pivot.columns)))
    ax.set_xticklabels([f"EF={c}" for c in pivot.columns])
    ax.set_yticks(range(len(pivot.index)))
    ax.set_yticklabels([f"S={s}" for s in pivot.index])
    ax.set_title("Harmonic-Mean TEPS Heatmap (Scale × Edge Factor)")
    plt.colorbar(im, ax=ax, label="TEPS")
    fig.tight_layout()
    fig.savefig(outdir / "teps_heatmap.png")
    plt.close(fig)
    print("  Saved teps_heatmap.png")


def plot_run_timeline(runs: pd.DataFrame, outdir: Path):
    """TEPS per BFS run index for each config — shows warm-up / variance."""
    if runs.empty:
        return

    configs = runs.assign(
        config=runs.apply(lambda r: label(int(r.scale), int(r.edge_factor)), axis=1)
    )
    unique = sorted(configs["config"].unique())
    n = len(unique)

    fig, axes = plt.subplots(n, 1, figsize=(10, 3 * n), sharex=True)
    if n == 1:
        axes = [axes]

    for ax, cfg in zip(axes, unique):
        sub = configs[configs["config"] == cfg].sort_values("run_index")
        ax.plot(sub["run_index"], sub["teps"], marker=".", linewidth=0.8)
        ax.set_ylabel("TEPS")
        ax.set_title(f"Config {cfg}")
        sci(ax)

    axes[-1].set_xlabel("BFS Run Index")
    fig.suptitle("TEPS per BFS Run (Timeline)", fontsize=13, y=1.01)
    fig.tight_layout()
    fig.savefig(outdir / "teps_timeline.png", bbox_inches="tight")
    plt.close(fig)
    print("  Saved teps_timeline.png")
