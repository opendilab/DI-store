import cProfile
import pstats
import io
from functools import partial


def print_stats(prof, sortby='cumulative'):
    s = io.StringIO()
    ps = pstats.Stats(prof, stream=s).sort_stats(sortby)
    ps.print_stats()
    print(s.getvalue())


def profile_it(fn=None, count=1):
    if fn is None:
        return partial(profile_it, count=count)

    current_count = 0
    prof = cProfile.Profile()

    def wrapped(*args, **kwargs):
        nonlocal current_count, prof
        try:
            return prof.runcall(fn, *args, **kwargs)
        finally:
            current_count += 1
            if current_count == count:
                print_stats(prof)
                current_count = 0
                prof = cProfile.Profile()

    return wrapped
