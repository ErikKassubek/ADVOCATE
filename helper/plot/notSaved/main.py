import matplotlib.pyplot as plt

plt.rcParams.update({
        'font.size': 22,
        'axes.labelsize': 24,
        'axes.titlesize': 24,
        'xtick.labelsize': 20,
        'ytick.labelsize': 20,
        'legend.fontsize': 20,
        'legend.title_fontsize': 20,
    })

x = list(range(2, 12, 1))

y1 = [80, 81] + [82] * 8  # found bugs

y2 = [5.27, 5.12, 5.39, 5.31, 5.18, 5.24, 5.35, 5.14, 5.33, 5.21]
y2.reverse()

print(x)
print(y1)

# Create figure and first axis
fig, ax1 = plt.subplots(figsize=(10, 5))

# Number Indicated bugs
l1, = ax1.plot(x, y1, 'r.-', label='Found bugs')

ax1.set_xlabel('SC length')
ax1.set_ylabel('Found bugs', color='r')
ax1.tick_params(axis='y', labelcolor='r')
ax1.set_ylim(70, 85)

ax1.axhline(y=83, color='r', linestyle='--')
ax1.set_yticks(range(71, 85, 2))


ax2 = ax1.twinx()
ax2.set_ylim(4, 8)

l3, = ax2.plot(x, y2, 'bx-', label='Runtime overhead')
#     ax2.set_ylabel('runtime [min]', color='r')
ax2.set_ylabel('runtime overhead', color='b')
ax2.tick_params(axis='y', labelcolor='b')
ax2.set_yticks(range(4, 8))  #


# Set legend
# ax1.legend(loc='lower left')

plt.tight_layout()
plt.show()
