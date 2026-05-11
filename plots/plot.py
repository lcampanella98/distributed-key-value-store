import matplotlib.pyplot as plt

# -----------------------------
# Experiment Data
# -----------------------------

scenarios = {
    "No Repair": {
        "stale_read_pct": 3.38,
        "remaining_pct": {
            "100ms": 100.00,
            "1s": 100.00,
            "2s": 100.00,
            "6s": 100.00,
        },
        "window": "9.99s"
    },

    "Immediate Repair": {
        "stale_read_pct": 1.20,
        "remaining_pct": {
            "100ms": 100.00,
            "1s": 100.00,
            "2s": 100.00,
            "6s": 100.00,
        },
        "window": "3.49s"
    },

    "Delayed Repair": {
        "stale_read_pct": 1.40,
        "remaining_pct": {
            "100ms": 96.41,
            "1s": 77.69,
            "2s": 57.77,
            "6s": 0.00,
        },
        "window": "4.79s"
    }
}

# -----------------------------
# Generate Plots
# -----------------------------

for scenario_name, data in scenarios.items():

    intervals = list(data["remaining_pct"].keys())

    # Convert:
    # (% stale reads) * (% stale reads remaining)
    # into percentage of TOTAL read requests
    remaining_of_total_reads = [
        data["stale_read_pct"] * (pct / 100.0)
        for pct in data["remaining_pct"].values()
    ]

    plt.figure(figsize=(8, 5))

    # plt.plot(
    #     intervals,
    #     remaining_of_total_reads,
    #     marker='o',
    #     linewidth=2
    # )
    x_values = [0, 0.3, 1, 2, 6]
    x_labels = ["0s", "100ms", "1s", "2s", "6s"]

    # Example initial y value
    initial_value = data["stale_read_pct"]

    y_values = [initial_value] + remaining_of_total_reads
    plt.plot(
        x_values,
        y_values,
        marker='o',
        linewidth=2
    )

    # Show readable labels instead of raw numeric values
    plt.xticks(x_values, x_labels)

    plt.title(f"{scenario_name} - Remaining Stale Reads Over Time")
    plt.xlabel("Time Interval")
    plt.ylabel("Remaining Stale Reads (% of Total Read Requests)")

    from matplotlib.ticker import PercentFormatter
    plt.gca().yaxis.set_major_formatter(PercentFormatter())
    # Add stale read window text
    plt.text(
        0.02,
        0.95,
        f"Stale Read Window: {data['window']}",
        transform=plt.gca().transAxes,
        verticalalignment='top',
        bbox=dict(boxstyle="round", alpha=0.2)
    )

    plt.grid(True)

    # Optional: tighten y-axis a bit
    plt.ylim(bottom=0, top=4)

    plt.tight_layout()
    # Save figure as PNG
    plt.savefig(
        f"{scenario_name.lower().replace(' ', '_')}.png",
        dpi=300,
        bbox_inches='tight'
    )
    # plt.show()