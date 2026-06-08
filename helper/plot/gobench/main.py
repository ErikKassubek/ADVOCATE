import matplotlib.pyplot as plt
from matplotlib.lines import Line2D
import random
import math

base = 600505
num = 83


def plotGraph1():
    x = range(2, 12)

    # prec
    allElemPrec = [96.38, 97.59, 98.79, 98.79, 98.79, 98.79, 97.59, 98.79, 98.79, 98.79] # all elements
    sameElemPrec = [79.51, 80.72, 81.92, 80.72, 81.92, 81.92, 81.92, 81.92, 81.92, 81.92] # same elements


    # overhead
    allElemOh = [5.47, 5.25, 5.45, 5.102, 5.49, 5.48, 5.40, 5.11, 5.37, 5.12] # all elements
    sameElemOh = [5.13, 5.07, 5.19, 5.26, 5.44, 5.43, 5.30, 5.38, 5.20, 5.42]  # same elements


    if True:
        fig, ax1 = plt.subplots()
        l1, = ax1.plot(x, allElemPrec, 'b^--')
        l2, = ax1.plot(x, sameElemPrec, 'bo-')
        ax1.set_xlabel('SC Length')
        ax1.set_ylabel('Found bugs [%]', color='b')
        ax1.tick_params(axis='y', labelcolor='b')
        ax1.set_ylim(0, 100)

    if True:
        # Create second axis
        ax2 = ax1.twinx()
    #     fig, ax2 = plt.subplots()
        # Runtime
        l3, = ax2.plot(x, allElemOh, 'r^--', label='GoPie')
        l4, = ax2.plot(x, sameElemOh, 'ro-', label='GoPie+')
    #     ax2.set_ylabel('runtime [min]', color='r')
        ax2.set_ylabel('Avg. runtime overhead factor per run', color='r')
        ax2.tick_params(axis='y', labelcolor='r')
        ax2.set_ylim(4, 7)

    custom_lines = [
        Line2D([0], [0], color='black', marker='^',
               linestyle='--', label='All Type'),
        Line2D([0], [0], color='black', marker='o',
               linestyle='-',  label='Same Type'),
    ]

    # Set legend
    ax1.legend(handles=custom_lines, loc='lower left')

    plt.tight_layout()
    plt.show()

def plotGraph2():
    x = [10 * max(2, int(math.pow(10, math.sqrt(x)))) for x in range(11)]
    
    # prec
    elemSameOver = [x + 0.4 for x in [4.48, 4.57, 4.56, 5.08, 5.00, 5.04, 5.22, 5.93, 5.48, 6.63, 7.31]]
    elemDiffOver = [x + 0.4 for x in [4.45, 4.59, 4.71, 4.91, 5.19, 4.84, 5.06, 6.14, 5.77, 7.12, 7.61]]
    routOver = [x + 0.4 for x in [4.79, 5.21, 4.95, 5.33, 5.74, 5.90, 6.24, 6.77, 8.24, 8.94, 9.42]]

    # Create figure and first axis

    if True:
        fig, ax1 = plt.subplots()
        l1, = ax1.plot(x, elemSameOver, 'b^-', label="Elements (Same)")
        l2, = ax1.plot(x, elemDiffOver, 'ro-', label="Elements (Different)")
        l3, = ax1.plot(x, routOver, 'gx-', label="Routines")
        ax1.set_xlabel('Nr.')
        ax1.set_ylabel('Overhead')
        ax1.tick_params(axis='y')
        ax1.set_ylim(1, 11)

    ax1.legend()
    
    plt.tight_layout()
    plt.show()

def plotAll():
    # ==================================================
    # Data from Graph 1
    # ==================================================
    x1 = range(2, 12)

    allElemPrec = [96.38, 97.59, 98.79, 98.79, 98.79,
                   98.79, 97.59, 98.79, 98.79, 98.79]

    sameElemPrec = [79.51, 80.72, 81.92, 80.72, 81.92,
                    81.92, 81.92, 81.92, 81.92, 81.92]

    allElemOh = [5.47, 5.25, 5.45, 5.102, 5.49,
                 5.48, 5.40, 5.11, 5.37, 5.12]

    sameElemOh = [5.13, 5.07, 5.19, 5.26, 5.44,
                  5.43, 5.30, 5.38, 5.20, 5.42]

    # ==================================================
    # Data from Graph 2
    # ==================================================
    x2 = [10 * max(2, int(math.pow(10, math.sqrt(x))))
          for x in range(11)]

    elemSameOver = [x + 0.4 for x in
                    [4.48, 4.57, 4.56, 5.08, 5.00,
                     5.04, 5.22, 5.93, 5.48, 6.63, 7.31]]

    elemDiffOver = [x + 0.4 for x in
                    [4.45, 4.59, 4.71, 4.91, 5.19,
                     4.84, 5.06, 6.14, 5.77, 7.12, 7.61]]

    routOver = [x + 0.4 for x in
                [4.79, 5.21, 4.95, 5.33, 5.74,
                 5.90, 6.24, 6.77, 8.24, 8.94, 9.42]]

    # ==================================================
    # Figure
    # ==================================================
    fig, ax_left = plt.subplots(figsize=(9, 5))

    # -------------------------
    # Left axis (precision)
    # -------------------------
    l1, = ax_left.plot(x1, allElemPrec,
                       color='tab:blue', marker='^',
                       linestyle='--', linewidth=2,
                       label='Found Bugs (All Types)')

    l2, = ax_left.plot(x1, sameElemPrec,
                       color='tab:blue', marker='o',
                       linestyle='-', linewidth=2,
                       label='Found Bugs (Same Types)')

    ax_left.set_xlabel('SC Length')
    ax_left.set_ylabel('Found Bugs [%]', color='tab:blue')
    ax_left.tick_params(axis='y', colors='tab:blue')
    ax_left.set_ylim(0, 100)
    ax_left.spines['left'].set_color('tab:blue')

    # -------------------------
    # Right axis (SC overhead)
    # -------------------------
    ax_right = ax_left.twinx()

    r1, = ax_right.plot(x1, allElemOh,
                        color='tab:red', marker='^',
                        linestyle='--', linewidth=2,
                        label='SC Overhead (All Types)')

    r2, = ax_right.plot(x1, sameElemOh,
                        color='tab:red', marker='o',
                        linestyle='-', linewidth=2,
                        label='SC Overhead (Same Types)')

    ax_right.set_ylabel('Runtime Overhead Factor', color='tab:red')
    ax_right.tick_params(axis='y', colors='tab:red')
    ax_right.spines['right'].set_color('tab:red')
    ax_right.set_ylim(4, 11)

    # -------------------------
    # Top axis (Graph 2)
    # -------------------------
    ax_top = ax_right.twiny()

    t1, = ax_top.plot(x2, elemSameOver,
                      color='firebrick', marker='^',
                      linestyle='-.', linewidth=2,
                      label='Elements (Same)')

    t2, = ax_top.plot(x2, elemDiffOver,
                      color='indianred', marker='o',
                      linestyle='-.', linewidth=2,
                      label='Elements (Different)')

    t3, = ax_top.plot(x2, routOver,
                      color='darkred', marker='s',
                      linestyle=':', linewidth=2.5,
                      label='Routines')

    ax_top.set_xlabel('Nr. Elements / Routines', color='darkred')
    ax_top.tick_params(axis='x', colors='darkred')
    ax_top.spines['top'].set_color('darkred')

    ax_top.set_ylim(ax_right.get_ylim())

    # ==================================================
    # GRID
    # ==================================================
    ax_left.grid(True, linestyle=':', alpha=0.5)

    # ==================================================
    # LEGENDS (FIXED: explicit handles)
    # ==================================================

    left_legend = ax_left.legend(
        handles=[l1, l2],
        loc='center left',
        bbox_to_anchor=(0.02, 0.62),
        title='Found Bugs [%]',
        frameon=True
    )

    right_legend = ax_left.legend(
        handles=[r1, r2, t1, t2, t3],
        loc='center left',
        bbox_to_anchor=(0.02, 0.28),
        title='Runtime Overhead',
        frameon=True
    )

    ax_left.add_artist(left_legend)

    plt.tight_layout()
    plt.show()

def buildVal():
    text = """
0-0-4972-4972
1-0-7391-7391
2-0-1204-1204
3-0-1225-1225
4-0-1215-1215
5-0-1308-1308
6-0-1383-1383
7-0-1767-1767
8-0-2675-2675
9-0-5017-5017
10-0-10760-10760
"""


    rows = []

    for line in text.strip().splitlines():
        parts = line.strip().split("-")
        first_num = int(parts[0])
        last_num = int(int(parts[-1]))
        rows.append((first_num, last_num))

    rows.sort(key=lambda x: x[0])

    return [last_num for _, last_num in rows]


if __name__ == "__main__":
    # print(buildVal())
    # plotGraph1()
    # plotGraph2()
    plotAll()