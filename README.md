# Back end service developed for my master project at The University of Bergen

## Endpoints

### OPTIONS '/session'
Post a new Session

### GET '/sessions'
Get all sessions without datapoints and locations

### POST '/sessions'
Expects parameter "Finished" which is either 'false' or '1'
Get all session, without datapoints and locations,  with the given Finished value

### GET '/session/:id'
Get the session, without datapoints, with the given id

### GET '/fullsession/:id'
Get the session, with datapoints, with the given id

### PUT '/session/:id'
Update the session with the given id

### OPTIONS '/beacon'
Create a new Beacon

### OPTIONS '/beacon/delete'
Expects parameter "Id"
Delete the beacon with the given Id

### GET '/beacons'
Get all beacons

### POST '/sessionbeacon'
Create a new SessionBeacon

### POST '/map'
Upload a .png image file

### GET '/debug/drop'
Drop all tables
NOT IN USE

### GET '/debug/drop/sessions'
Drop Session and SessionBeacon tables
NOT IN USE
