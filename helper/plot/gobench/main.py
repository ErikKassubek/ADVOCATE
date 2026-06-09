import matplotlib.pyplot as plt
from matplotlib.lines import Line2D
import random
import math

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
    allElemPrec = [96.38, 97.59, 98.79, 98.79, 98.79, 98.79, 98.79, 98.79, 98.79, 98.79] # all elements
    sameElemPrec = [79.51, 80.72, 81.92, 81.92, 81.92, 81.92, 81.92, 81.92, 81.92, 81.92] # same elements


    # overhead
    allElemOh = [5.47, 5.25, 5.45, 5.102, 5.49, 5.48, 5.40, 5.11, 5.37, 5.12] # all elements
    sameElemOh = [5.13, 5.07, 5.19, 5.26, 5.44, 5.43, 5.30, 5.38, 5.20, 5.42]  # same elements


    if True:
        fig, ax1 = plt.subplots(figsize=(10, 6))
        l1, = ax1.plot(x, allElemPrec, 'b^--')
        l2, = ax1.plot(x, sameElemPrec, 'bo-')
        ax1.set_xlabel('SC Length')
        ax1.set_ylabel('Found bugs [%]', color='b')
        ax1.tick_params(axis='y', labelcolor='b')
        ax1.set_ylim(70, 100)

    if True:
        # Create second axis
        ax2 = ax1.twinx()
    #     fig, ax2 = plt.subplots()
        # Runtime
        l3, = ax2.plot(x, allElemOh, 'r^--', label='GoPie')
        l4, = ax2.plot(x, sameElemOh, 'ro-', label='GoPie+')
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
    ax1.legend(handles=custom_lines, loc='lower left')

  

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

    x = [10 * max(2, int(math.pow(10, math.sqrt(x)))) for x in range(11)]
    
    # prec
    # elemSameOver = [x + 0.4 for x in [4.48, 4.57, 4.56, 5.08, 5.00, 5.04, 5.22, 5.93, 5.48, 6.63, 7.31]]
    # elemDiffOver = [x + 0.4 for x in [4.45, 4.59, 4.71, 4.91, 5.19, 4.84, 5.06, 6.14, 5.77, 7.12, 7.61]]
    # routOver = [x + 0.4 for x in [4.79, 5.21, 4.95, 5.33, 5.74, 5.90, 6.24, 6.77, 8.24, 8.94, 9.42]]
    elemSameOver = [4.48, 4.57, 4.56, 5.08, 5.00, 5.04, 5.22, 5.93, 5.48, 6.63, 7.31]
    elemDiffOver = [4.45, 4.59, 4.71, 4.91, 5.19, 4.84, 5.06, 6.14, 5.77, 7.12, 7.61]
    routOver = [4.79, 5.21, 4.95, 5.33, 5.74, 5.90, 6.24, 6.77, 8.24, 8.94, 9.42]
    

    # Create figure and first axis

    if True:
        fig, ax1 = plt.subplots(figsize=(10, 6))
        l1, = ax1.plot(x, elemSameOver, 'b^-', label="Elements (Same)")
        l2, = ax1.plot(x, elemDiffOver, 'ro-', label="Elements (Different)")
        l3, = ax1.plot(x, routOver, 'gx-', label="Routines")
        ax1.set_xlabel('Nr.')
        ax1.set_ylabel('Runtime overhead')
        ax1.tick_params(axis='y')
        ax1.set_ylim(1, 11)

    ax1.legend(loc='lower left')

    
    plt.tight_layout()
    plt.show()


if __name__ == "__main__":
    # print(buildVal())
    plotGraph1()
    # plotGraph2()
