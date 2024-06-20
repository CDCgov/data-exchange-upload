# DEX Pipeline Reporter

This is an automation tool for regularly monitoring file activity in DEX.  Its goal is to allow DEX maintainers to quickly
see the history of file activity, detect any anomalous events on those files or the messages within, and alert the team 
if necessary.

## Checks

Here is a list of the data stores that this tool checks to get a status of files and messages over a given time range:

- Tus Upload Storage Container - This will show how many uploads were attempted and completed.
- Checkpoint Delivery Containers (edav and routing) - The sum of these files should equal the total files uploaded.  If
there is a difference, this tool shall perform further inspection to record any missing files.

## Reporting Schedule

This tool shall report on DEX twice a day.  This frequency shall be configurable.

## Report History

This tool shall store reports either in CSV or database tables.  This is to enable retroactive auditing.

## Recovery

In the event of an anomaly, this tool shall alert DEX maintainers.  At that point, maintainers will be given a list of 
unresolved anomalies.  Each anomaly will include the upload ID associated with it.  This will allow a maintainer to replay
the file in order to resolve the anomaly.

## Future Improvements

- Perform checks on message counts.  Need to figure out what our expected message total will be for a given time range.
- Get the expected total count from programs to represent the true expected total count
- Automatic anomaly resolution via automatic file replay