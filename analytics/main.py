"""
graph500_analysis.py
--------------------
Reads output/results.jsonl produced by the Go BFS benchmark and generates
plots for TEPS, BFS time, and edge traversal across scale / edge-factor configs.

Usage:
    python graph500_analysis.py                        # uses output/results.jsonl
    python graph500_analysis.py --input path/to/results.jsonl --outdir plots/
"""

import argparse
import json
from pathlib import Path

import matplotlib
import matplotlib.pyplot as plt
import pandas as pd

from plots import (
    plot_construct_time,
    plot_nedge_vs_scale,
    plot_run_timeline,
    plot_teps_box,
    plot_teps_heatmap,
    plot_teps_vs_scale,
    plot_time_vs_scale,
)


matplotlib.use("Agg")
plt.rcParams.update(
    {
        "figure.dpi": 150,
        "font.size": 11,
        "axes.grid": True,
        "grid.alpha": 0.3,
        "axes.spines.top": False,
        "axes.spines.right": False,
    }
)


def load_jsonl(path: str) -> tuple[pd.DataFrame, pd.DataFrame]:
    """Return (runs_df, summaries_df) parsed from a JSONL results file."""
    runs, summaries = [], []
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            obj = json.loads(line)
            if obj.get("type") == "run":
                runs.append(obj)
            elif obj.get("type") == "summary":
                summaries.append(obj)

    runs_df = pd.DataFrame(runs) if runs else pd.DataFrame()
    summ_df = pd.DataFrame(summaries) if summaries else pd.DataFrame()
    return runs_df, summ_df


# ── Summary table


def print_summary_table(summ: pd.DataFrame):
    cols = [
        "scale",
        "edge_factor",
        "nodes",
        "total_edges",
        "teps_harmonic_mean",
        "time_median_s",
        "construct_time_s",
    ]
    cols = [c for c in cols if c in summ.columns]
    print("\nSummary Table ")
    print(summ[cols].sort_values(["scale", "edge_factor"]).to_string(index=False))
    print()


def main():
    parser = argparse.ArgumentParser(description="Graph500 BFS benchmark analyser")
    parser.add_argument(
        "--input", default="output/results.jsonl", help="Path to results JSONL file"
    )
    parser.add_argument(
        "--outdir", default="plots", help="Directory to write PNG plots into"
    )
    args = parser.parse_args()

    outdir = Path(args.outdir)
    outdir.mkdir(parents=True, exist_ok=True)

    print(f"Loading {args.input} ...")
    runs, summ = load_jsonl(args.input)
    print(f"  {len(runs)} run records, {len(summ)} summary records loaded.")

    if summ.empty:
        print(
            "No summary records found — cannot produce most plots. "
            "Check that the benchmark wrote output correctly."
        )
        return

    print_summary_table(summ)

    print("Generating plots ...")
    plot_teps_vs_scale(summ, outdir)
    plot_time_vs_scale(summ, outdir)
    plot_nedge_vs_scale(summ, outdir)
    plot_construct_time(summ, outdir)
    plot_teps_heatmap(summ, outdir)
    plot_teps_box(runs, outdir)
    plot_run_timeline(runs, outdir)

    print(f"\nAll plots saved to '{outdir}/'")


if __name__ == "__main__":
    main()
