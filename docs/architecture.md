## Basic architecture
The webserver is centered aroud a `webserverStorage` object which contains a read/write lock for avoiding race conditions on operations and two maps of maps which organize the actual content of the objects and some metadata about the objects as well.

```
contentStorage  map[string]map[string]string // map of bucket->hash->actual body
metadataStorage map[string]map[string]string // map of bucket->id->sha256 hash of body
```

These two maps of maps exist because of the deduplication requirement. The `contentStorage` map is a map of bucketID to a hash of the content, which in turn is the key for the actual content. The `metadataStorage` map is a map of bucketID's to the objectID (supplied by the user) to a SHA256 hash of the content. I used SHA256 because it's basic and well known. A more secure application would use something more sophisticated and salts.

## PUT/GET/DELETE
The three operations modify the maps as in order to keep everything in order.

### PUT
When a valid request comes in, hash the request body, and store the resulting hash in `metadataStorage[bucket][id]`. Store the actual content that was hashed in `contentStorage[bucket][hash]`

### GET
To access content that exists, first check the metadata to see if it's actually been stored with the user supplied parameters. If so, then query the hash accessed from `metadataStorage[bucket][id]` in `contentStorage[bucket][hash]`.

### DELETE
This is the most complicated operation because you can't simply delete things from both maps. When a request comes in, start by deleting the hash from `metadata[bucket][id]`, since that is no longer going to be used. Then check to see if that hash exists in any other bucket by iterating through `metadataStorage[bucket]`, which looks at all the existing `id` values. If the hash exists somewhere, then don't update the `contentStorage` map. If it doesn't exist, then you can delete the content from `contentStorage[bucket][hash]`.

## Assumptions
The requirements indicated that the content for each request is `simply the text of an HTTPrequest.`. I chose to use the request body rather than a full dump because the likelihood of a collision would be far less if you included timestamps and headers, meaning it would be harder to test.

Deduplication per bucket means that we wont have identical items within a bucket. We could have identical items across buckets, however.

## Complexity
The delete operation has a complexity of * *O(N)* * given that it has to iterate over each object in a bucket. I considered this to be ok for now, but if this required some massive scale it would be worth implementing some sort of reference counting solution.