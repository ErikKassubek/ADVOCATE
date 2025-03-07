# End routine

When a routine is ended, an element is added to the trace.
The element is not added if the routine is ended because of a panic in another routine or because the main routine terminated.
# Trace element

The basic form of the trace element is

```
E,[t]
```

where `E` identifies the element as a routine end element. The following
fields are

- [t] $\in\mathbb N: This is the value of the global counter when the routine ended
