Cheesecake
----------

Cheesecake is an evidence-based scouting and statistics approach to
the *FIRST* Robotics Competition. It has two main goals:

1. To use existing data on the results of *FIRST* competitions to
   create measurably accurate prediction metrics.
2. To facilitate a mixed-methods (quantitative plus qualitative)
   approach to FRC scouting specific to each game.

Cheesecake takes an empirical, evidence-based approach to how it
handles FRC data.

<!--
Using Cheesecake, we hope to be able to answer the following questions:

+ How likely are teams to qualify for championships?
+ 
-->

Motivation
==========

Scouting is a large commitment for a team. At most competitions we
attend, we usually allocate a significant amount of team resources to
ensure we have as much data as possible on each robot at the
competition. This information typically goes into a binder and is used
by our scouting team to determine the best robots to pick during the
alliance selections and the optimal strategies to play against
individual robots.

The goal of Cheesecake is to ensure that scouting information is
transformed into useful metrics. It draws inspiration (and some
models) from other types of sports analytics, statistics, and previous
related systems.

Tools
=====

Since this is a side project, we wanted to use this as an opportunity
to try out some new tools. The back-end is currently being developed
using Sanic, an asynchronous Python web server. The front-end tools
are still being identified, but we'll probably end up using React
or something similar.

Design
======

Coming soon. Based on some early sketches, Cheesecake will likely contain



1. A web interface for scouting data entry
2. Scrapers for gathering third-party data (scores, breakdowns, etc.)
3. A system for statistical analysis and visualization
4. Match and rank predictions


Installation
============

It is recommended that you use a virtual environment when installing
and setting up Cheesecake.

```python
pip install -r requirements.txt
```
