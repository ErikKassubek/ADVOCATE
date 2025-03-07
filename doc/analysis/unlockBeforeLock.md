# Unlock before lock

If a not locked mutex is unlocked the program panics.
This is possible, if we can reorder lock and unlock operations, such that
at some point there are more executed unlock than lock operations.

This is equal to the analysis scenario done before add, such that
lock operations equal add operations and unlock operations equal done operations.
The detection of such situations is therefore done equivalent to the
done before add analysis scenario.

The analysis for this is fairly slow. For the done before add detection,
this has a rather small impact, because the number of adds and dones on one
wait group is normally rather small. For mutexes this is not the case. For this
reason we assume the following. We assume that an unlock before lock can only
happen, if an unlock operation on a mutex happens in a different routine than
the corresponding lock operations. Before we run the analysis for unlock before lock,
we therefore check if this is the case for the mutex, and otherwise skip the
further analysis.