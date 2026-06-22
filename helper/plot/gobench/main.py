import matplotlib.pyplot as plt
from matplotlib.lines import Line2D
import random
import math
import numpy as np

base = 600505
num = 83


def plotGraph1():
    plt.rcParams.update({
        'font.size': 22,
        'axes.labelsize': 24,
        'axes.titlesize': 24,
        'xtick.labelsize': 20,
        'ytick.labelsize': 20,
        'legend.fontsize': 20,
        'legend.title_fontsize': 20,
    })
    
    x = range(2, 12)

    # prec
    allElemPrec = [80, 81, 82, 82, 82, 82, 82, 82, 82, 82] # all elements
    sameElemPrec = [66, 67, 68, 68, 68, 68, 68, 68, 68, 68] # same elements


    # overhead
    allElemOh = [5.47, 5.25, 5.45, 5.102, 5.49, 5.48, 5.40, 5.11, 5.37, 5.12] # all elements
    sameElemOh = [5.13, 5.07, 5.19, 5.26, 5.44, 5.43, 5.30, 5.38, 5.20, 5.42]  # same elements


    if True:
        fig, ax1 = plt.subplots(figsize=(10, 6))
        l1, = ax1.plot(x, allElemPrec, 'b^--')
        l2, = ax1.plot(x, sameElemPrec, 'bo-')
        ax1.set_xlabel('SC Length')
        ax1.set_ylabel('Found bugs', color='b')
        ax1.tick_params(axis='y', labelcolor='b')
        ax1.set_ylim(58, 84)

        ax1.axhline(y=83, color='b', linestyle=':', linewidth=2, label='y=83')

    if True:
        # Create second axis
        ax2 = ax1.twinx()
    #     fig, ax2 = plt.subplots()
        # Runtime
        l3, = ax2.plot(x, allElemOh, 'r^--')
        l4, = ax2.plot(x, sameElemOh, 'ro-')
    #     ax2.set_ylabel('runtime [min]', color='r')
        ax2.set_ylabel('Runtime overhead', color='r')
        ax2.set_yticks([4, 4.5, 5, 5.5, 6])
        ax2.tick_params(axis='y', labelcolor='r')
        ax2.set_ylim(4, 6)

    custom_lines = [
        Line2D([0], [0], color='black', marker='^',
               linestyle='--', label='All Type'),
        Line2D([0], [0], color='black', marker='o',
               linestyle='-',  label='Same Type'),
    ]

    # Set legend
    ax1.legend(handles=custom_lines, loc='lower right')

  

    plt.tight_layout()
    plt.show()

def plotGraph2():
    plt.rcParams.update({
        'font.size': 22,
        'axes.labelsize': 24,
        'axes.titlesize': 24,
        'xtick.labelsize': 20,
        'ytick.labelsize': 20,
        'legend.fontsize': 20,
        'legend.title_fontsize': 20,
    })

    # x-axis (approx. from figure)
    x = np.array([0, 0.05, 0.1, 0.2, 0.4, 0.7, 1.1, 1.6])

    # Upper curves (solid lines)
    ops_upper = [4.9, 5.0, 5.1, 5.5, 6.2, 7.0, 8.2, 9.8]
    routines_upper = [4.8, 4.9, 5.0, 5.4, 5.7, 6.4, 7.1, 8.2]

    # Lower curves (dashed lines)
    ops_lower = [2.1, 2.2, 2.3, 2.5, 2.9, 3.4, 4.1, 5.1]
    routines_lower = [2.3, 2.35, 2.4, 2.6, 2.8, 3.2, 3.5, 4.1]

    # Plot
    fig, ax = plt.subplots(figsize=(10, 5))

    # Solid lines (upper)
    ax.plot(x, ops_upper, 'bo-', label='Operations')
    ax.plot(x, routines_upper, 'g*-', label='Routines')

    # Dashed lines (lower)
    ax.plot(x, ops_lower, 'b--o')
    ax.plot(x, routines_lower, 'g--*')

    # Labels
    ax.set_ylabel('Overhead')
    ax.set_xlabel('Nr. Ops [M], Routines [0.1M]')

    # Legend
    ax.legend(loc='upper left')

    plt.tight_layout()
    plt.show()

if __name__ == "__main__":
    # print(buildVal())
    # plotGraph1()
    plotGraph2()
