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

    x = [2] + [15 * max(1, int(math.pow(10, math.sqrt(x)))) for x in range(10)]

    elemSameOver  = [4.88, 4.90, 4.96, 5.18, 5.34, 5.44, 5.62, 6.08, 6.54, 7.03, 7.71]
    elemDiffOver  = [4.85, 4.96, 5.01, 5.11, 5.29, 5.38, 5.46, 6.24, 6.88, 7.52, 8.01]
    routOver      = [5.19, 5.28, 5.45, 5.73, 6.14, 6.30, 6.64, 7.17, 8.02, 9.12, 9.82]

    elemSameOver2 = [2.47, 2.48, 2.51, 2.71, 2.98, 2.94, 3.24, 3.45, 4.07, 4.40, 4.63]
    elemDiffOver2 = [2.43, 2.50, 2.66, 2.93, 3.23, 3.41, 3.49, 3.61, 3.94, 4.25, 4.72]
    routOver2     = [2.48, 2.51, 2.78, 2.94, 3.29, 3.49, 3.70, 4.33, 5.10, 5.78, 6.41]



    

    # Create figure and first axis

    if True:
        fig, ax1 = plt.subplots(figsize=(10, 6))
        l1, = ax1.plot(x, elemSameOver, 'b^-', label="Elements (Same)")
        l2, = ax1.plot(x, elemDiffOver, 'ro-', label="Elements (Different)")
        l3, = ax1.plot(x, routOver, 'gx-', label="Routines")

        l4, = ax1.plot(x, elemSameOver2, 'b^--')
        l5, = ax1.plot(x, elemDiffOver2, 'ro--')
        l6, = ax1.plot(x, routOver2, 'gx--')

        ax1.set_xlabel('Nr.')
        ax1.set_ylabel('Runtime overhead')
        ax1.tick_params(axis='y')
        ax1.set_ylim(1, 12)

    ax1.legend(loc='upper left')

    
    plt.tight_layout()
    plt.show()


if __name__ == "__main__":
    # print(buildVal())
    plotGraph1()
    # plotGraph2()
