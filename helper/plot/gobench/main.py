import matplotlib.pyplot as plt
from matplotlib.lines import Line2D

x = [2, 5, 10, 25, 50, 75, 100]

# gopie
y1_1 = [60, 60, 60, 59, 62, 62, 60]  # found bugs
y1_2 = [56, 55, 55, 54, 58, 58, 56]  # found progs with bugs
y1_3 = [7.15, 17.53, 17.45, 19.55, 22.63, 32.47, 27.6]  # runtime in s


# gopieplus
y2_1 = [62, 66, 60, 66, 62, 66, 66]        # found bugs
y2_2 = [58, 62, 55, 62, 58, 62, 62]        # found progs with bugs
y2_3 = [6.78, 6.71, 8.11, 6.93, 6.68, 6.97, 8.22]        # runtime in s


if True:
    # Create figure and first axis
    fig, ax1 = plt.subplots()

    if True:
        # Number Indicated bugs
        l1, = ax1.plot(x, y1_1, 'b^--')
        l2, = ax1.plot(x, y2_1, 'bo-')
        ax1.set_xlabel('Max number fuzzing runs per test')
        ax1.set_ylabel('Found bugs', color='b')
        ax1.tick_params(axis='y', labelcolor='b')
        ax1.set_ylim(45, 68)
    else:
        # Number tests with bugs
        l1, = ax1.plot(x, y1_2, 'b^--')
        l2, = ax1.plot(x, y2_2, 'bo-')
        ax1.set_xlabel('Max number fuzzing runs per test')
        ax1.set_ylabel('Tests where bugs were found', color='b')
        ax1.tick_params(axis='y', color='b')
        ax1.set_ylim(30, 65)

if False:
    # Create second axis
    ax2 = ax1.twinx()
#     fig, ax2 = plt.subplots()
    # Runtime
    l3, = ax2.plot(x, y1_3, 'r^--', label='GoPie')
    l4, = ax2.plot(x, y2_3, 'ro-', label='GoPie+')
#     ax2.set_ylabel('runtime [min]', color='r')
    ax2.set_ylabel('runtime [min]')
    ax2.tick_params(axis='y', labelcolor='r')
    ax2.set_ylim(0, 50)

custom_lines = [
    Line2D([0], [0], color='black', marker='^',
           linestyle='--', label='GoPie'),
    Line2D([0], [0], color='black', marker='o',
           linestyle='-',  label='GoPie+'),
]

# Set legend
ax1.legend(handles=custom_lines, loc='upper left')

plt.tight_layout()
plt.show()
