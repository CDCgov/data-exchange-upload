## Metadata Management Service (MMS) API Documentation
 
### Overview
 
The Metadata Management Service (MMS) is designed to handle metadata for the DEX ecosystem. This document provides an overview of the available endpoints, their functionalities, and usage examples.
 
#### Base URL
 
All API endpoints are relative to the base URL:

local environment - `http://localhost:8000`\
DEV environment - `https://dev-portal-mms.azurewebsites.net`
 
### Endpoints
 
#### Metrics
 
##### Verify Database Health
 
- **Endpoint**: `/health`
- **Method**: `GET`
- **Description**: Verify the database connection status.
- **Example Request**:
  ```
  GET http://localhost:8000/health
  ```
 
#### Entities
 
Entities represent the organizations and programs interacting with the system.
 
##### Create New Entity
 
- **Endpoint**: `/entities`
- **Method**: `POST`
- **Description**: Create a new entity.
- **Example Request**:
  ```
  POST http://localhost:8000/entities
  ```
 
##### Fetch All Entities
 
- **Endpoint**: `/entities`
- **Method**: `GET`
- **Description**: Fetch all entities.
- **Example Request**:
  ```
  GET http://localhost:8000/entities
  ```
 
##### Fetch Entity by ID
 
- **Endpoint**: `/entities/{{entity_id}}`
- **Method**: `GET`
- **Description**: Fetch an entity by its ID.
- **Example Request**:
  ```
  GET http://localhost:8000/entities/1
  ```
 
##### Delete Entity by ID
 
- **Endpoint**: `/entities/{{entity_id}}`
- **Method**: `DELETE`
- **Description**: Delete an entity by its ID.
- **Example Request**:
  ```
  DELETE http://localhost:8000/entities/1
  ```
 
#### Programs
 
##### Create Program for an Entity
 
- **Endpoint**: `/entities/{{entity_id}}/programs`
- **Method**: `POST`
- **Description**: Create a program for an entity.
- **Example Request**:
  ```
  POST http://localhost:8000/entities/1/programs
  ```
 
##### Fetch All Entity Programs
 
- **Endpoint**: `/entities/{{entity_id}}/programs`
- **Method**: `GET`
- **Description**: Fetch all programs for an entity.
- **Example Request**:
  ```
  GET http://localhost:8000/entities/1/programs
  ```
 
##### Fetch Entity Program by ID
 
- **Endpoint**: `/entities/{{entity_id}}/programs/{{program_id}}`
- **Method**: `GET`
- **Description**: Fetch a program by its ID.
- **Example Request**:
  ```
  GET http://localhost:8000/entities/1/programs/1
  ```
 
##### Delete Program by ID
 
- **Endpoint**: `/entities/{{entity_id}}/programs/{{program_id}}`
- **Method**: `DELETE`
- **Description**: Delete a program by its ID.
- **Example Request**:
  ```
  DELETE http://localhost:8000/entities/1/programs/1
  ```
 
#### Authorization Groups
 
Authorization Groups allow for more granular control of access to certain datastreams.
 
##### Create New Group for an Entity
 
- **Endpoint**: `/entities/{{entity_id}}/groups`
- **Method**: `POST`
- **Description**: Create a new group for an entity.
- **Example Request**:
  ```
  POST http://localhost:8000/entities/1/groups
  ```
 
##### Fetch All Groups in an Entity
 
- **Endpoint**: `/entities/{{entity_id}}/groups`
- **Method**: `GET`
- **Description**: Fetch all groups in an entity.
- **Example Request**:
  ```
  GET http://localhost:8000/entities/1/groups
  ```
 
##### Fetch Group in an Entity by ID
 
- **Endpoint**: `/entities/{{entity_id}}/groups/{{group_id}}`
- **Method**: `GET`
- **Description**: Fetch a group in an entity by its ID.
- **Example Request**:
  ```
  GET http://localhost:8000/entities/1/groups/1
  ```
 
##### Delete Group by ID
 
- **Endpoint**: `/entities/{{entity_id}}/groups/{{group_id}}`
- **Method**: `DELETE`
- **Description**: Delete a group by its ID.
- **Example Request**:
  ```
  DELETE http://localhost:8000/entities/1/groups/1
  ```
 
#### Datastreams
 
##### Create New Datastream
 
- **Endpoint**: `/datastreams`
- **Method**: `POST`
- **Description**: Create a new datastream.
- **Example Request**:
  ```
  POST http://localhost:8000/datastreams
  ```
 
##### Fetch All Datastreams
 
- **Endpoint**: `/datastreams`
- **Method**: `GET`
- **Description**: Fetch all datastreams.
- **Example Request**:
  ```
  GET http://localhost:8000/datastreams
  ```
 
##### Fetch Datastream by ID
 
- **Endpoint**: `/datastreams/{{datastream_id}}`
- **Method**: `GET`
- **Description**: Fetch a datastream by its ID.
- **Example Request**:
  ```
  GET http://localhost:8000/datastreams/1
  ```
 
##### Delete Datastream by ID
 
- **Endpoint**: `/datastreams/{{datastream_id}}`
- **Method**: `DELETE`
- **Description**: Delete a datastream by its ID.
- **Example Request**:
  ```
  DELETE http://localhost:8000/datastreams/1
  ```
 
#### Authorization Groups for Datastreams
 
##### Fetch All Subgroups in a Datastream
 
- **Endpoint**: `/datastreams/{{datastream_id}}/authorizationgroups`
- **Method**: `GET`
- **Description**: Fetch all subgroups in a datastream.
- **Example Request**:
  ```
  GET http://localhost:8000/datastreams/1/authorizationgroups
  ```
 
##### Fetch Subgroups Data by ID
 
- **Endpoint**: `/datastreams/{{datastream_id}}/authorizationgroups/{{auth_group_id}}`
- **Method**: `GET`
- **Description**: Fetch subgroups data by ID.
- **Example Request**:
  ```
  GET http://localhost:8000/datastreams/1/authorizationgroups/1
  ```
 
#### Routes
 
##### Create New Route for a Datastream
 
- **Endpoint**: `/datastreams/{{datastream_id}}/routes`
- **Method**: `POST`
- **Description**: Create a new route for a datastream.
- **Example Request**:
  ```
  POST http://localhost:8000/datastreams/1/routes
  ```
 
##### Fetch All Routes in a Datastream
 
- **Endpoint**: `/datastreams/{{datastream_id}}/routes`
- **Method**: `GET`
- **Description**: Fetch all routes in a datastream.
- **Example Request**:
  ```
  GET http://localhost:8000/datastreams/1/routes
  ```
 
##### Fetch Route in Datastream by ID
 
- **Endpoint**: `/datastreams/{{datastream_id}}/routes/{{datastreamroute_id}}`
- **Method**: `GET`
- **Description**: Fetch a route in a datastream by its ID.
- **Example Request**:
  ```
  GET http://localhost:8000/datastreams/1/routes/1
  ```
 
##### Delete Datastream Route by ID
 
- **Endpoint**: `/datastreams/{{datastream_id}}/routes/{{datastreamroute_id}}`
- **Method**: `DELETE`
- **Description**: Delete a datastream route by its ID.
- **Example Request**:
  ```
  DELETE http://localhost:8000/datastreams/1/routes/1
  ```
 
#### Manifests
 
##### Create Manifest for Datastream Route
 
- **Endpoint**: `/datastreams/{{datastream_id}}/routes/{{datastreamroute_id}}/manifests`
- **Method**: `POST`
- **Description**: Create a manifest for a datastream route.
- **Example Request**:
  ```
  POST http://localhost:8000/datastreams/dextesting/routes/testevent1/manifests
  ```
 
##### Fetch All Manifests for Datastream Route
 
- **Endpoint**: `/datastreams/{{datastream_id}}/routes/{{datastreamroute_id}}/manifests`
- **Method**: `GET`
- **Description**: Fetch all manifests for a datastream route.
- **Example Request**:
  ```
  GET http://localhost:8000/datastreams/dextesting/routes/testevent1/manifests
  ```
 
##### Delete Manifest by ID
 
- **Endpoint**: `/datastreams/{{datastream_id}}/routes/{{datastreamroute_id}}/manifests/{{manifest_id}}`
- **Method**: `DELETE`
- **Description**: Delete a manifest by its ID.
- **Example Request**:
  ```
  DELETE http://localhost:8000/datastreams/dextesting/routes/testevent1/manifests/1
  ```

## Sender Manifest Retrieval and Caching Strategy
 
### Overview
 
The `APIConfigLoader` uses a caching strategy to enhance performance and reduce the load on the API by storing configuration data in Redis. The cache ensures that frequently accessed configurations are quickly retrieved from memory rather than making repeated API calls.
 
### Caching Mechanism
 
#### Cache Key
 
- Each configuration path is associated with a unique cache key.
- The cache key is constructed by concatenating a prefix (`"config:"`) with the configuration path.
 
#### Example
 
For a configuration path `"/config/path"`, the cache key would be `"config:/config/path"` (i.e. `"config: celr-csv"`).
 
### Cache Retrieval
 
1. **Attempt to Retrieve from Cache**: When `LoadConfig` is called, the first step is to try to retrieve the configuration data from Redis using the constructed cache key.
   - If the data is found in the cache, it is returned immediately.
   - If the data is not found (`redis.Nil`), the method proceeds to fetch the data from the API.
   - If an error other than a cache miss occurs, it is returned.
 
#### Example Code
 
```go
cachedConfig, err := l.RedisClient.Get(ctx, cacheKey).Bytes()
if err == nil {
    return cachedConfig, nil
}
if err != redis.Nil {
    return nil, err
}
```
 
### Cache Miss Handling
 
2. **Retrieve from API on Cache Miss**: If the configuration data is not found in the cache:
   - An HTTP GET request is made to the API using the provided base URL and path.
   - If the API response is successful (`HTTP 200`), the data is read from the response body.
   - If the API response is not found (`HTTP 404`), an error is returned indicating that the configuration was not found.
   - For any other unsuccessful responses, a generic error is returned.
 
#### Example Code
 
```go
url := l.BaseURL + path
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
if err != nil {
    return nil, err
}
 
resp, err := l.HTTPClient.Do(req)
if err != nil {
    return nil, err
}
defer resp.Body.Close()
 
if resp.StatusCode != http.StatusOK {
    if resp.StatusCode == http.StatusNotFound {
        return nil, errors.New("configuration not found")
    }
    return nil, errors.New("failed to retrieve configuration")
}
 
configData, err := io.ReadAll(resp.Body)
if err != nil {
    return nil, err
}
```
 
### Storing in Cache
 
3. **Store in Cache**: After successfully retrieving the configuration data from the API:
   - The data is stored in Redis with the constructed cache key.
   - The data is stored with a specified time-to-live (TTL) duration to ensure it is cached for an appropriate period before expiring.
 
#### Example Code
 
```go
err = l.RedisClient.Set(ctx, cacheKey, configData, l.CacheTTL).Err()
if err != nil {
    return nil, err
}
```
 
### Cache Invalidation
 
4. **Cache Invalidation**: The `InvalidateCache` method provides functionality to invalidate (delete) the cache for a specific configuration path:
   - This is useful when the configuration data is updated and the cache needs to be refreshed.
   - The method constructs the cache key and deletes the corresponding entry from Redis.
 
#### Example Code
 
```go
cacheKey := "config:" + path
return l.RedisClient.Del(ctx, cacheKey).Err()
```
 
## Cache TTL (Time-to-Live)
 
- The cache TTL is a duration specified during the initialization of the `APIConfigLoader`.
- It determines how long the cached data is valid before it expires and needs to be refreshed from the API.
- A suitable TTL balances between performance (reducing API calls) and data freshness.
 
#### Example
 
```go
loader := api.NewAPIConfigLoader("https://api.example.com", "localhost:6379", 5*time.Minute)
```
 
In this example, the cached configuration data is valid for 5 minutes.
 
## Summary
 
- **Cache Key Construction**: Uses a prefix and the configuration path.
- **Cache Retrieval**: Attempts to get data from Redis first; on a miss, fetches from the API.
- **Cache Miss Handling**: Fetches data from the API and handles various HTTP responses.
- **Storing in Cache**: Saves the retrieved data in Redis with a TTL.
- **Cache Invalidation**: Deletes specific cache entries when data is updated.
- **Cache TTL**: Configurable duration for how long cached data remains valid.
 
This caching strategy ensures that configuration data is efficiently retrieved and kept up-to-date, balancing performance with data freshness.