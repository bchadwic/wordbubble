
# WordBubble

REST queue for sending and receiving information

## API Reference (v0.0.1)

### Sign up to WordBubble

```
POST /signup
```
##### Request Body
```json
{
    "username": "bchadwic",
    "email": "benchadwick87@gmail.com",
    "password": "Wordbubble123!"
}
```
##### Response Body
```
${token}
```

#### Parameters
|     Field     |  Description  |  Constraints |Required |
| ------------- | ------------- | ------------ | ---- |
|   `username`  | name used to identify user  | `1-40` characters of `a-z`, `1-9` or `_` | ✓ |
| `email`  | email used to identify user  | a valid email address | ✓ |
| `password`  | key to get a token  | `>6` characters of an uppercase, a lowercase, a number, and a symbol | ✓ |


### Retrieve a token

```
POST /token
```
##### Request Body
```json
{
    "username": "bchadwic",
    "password": "Wordbubble123!"
}
```
##### Response Body
```
${token}
```
#### Parameters
|     Field     |  Description  |  Constraints | Required |
| ------------- | ------------- | ------------ | ---- |
|   `username`  | name used to identify user  | `1-40` characters of `a-z`, `1-9` or `_` | ✓ |
| `password`  | key to get a token  | `>6` characters of an uppercase, a lowercase, a number, and a symbol | ✓ |

### Push a new WordBubble

```
POST /push
```
##### Request Body
```json
{
    "text": "Hello World!"
}
```
##### Response Body
```
thank you!
```
#### Parameters
|     Field     |  Description  |  Constraints | Required |
| ------------- | ------------- | ------------ | ---- |
| `text`  | text to be added to the queue | `1-255` characters | ✓ |

#### Headers
|     Field     |  Description  |  Constraints | Required |
| ------------- | ------------- | ------------ | ---- |
| `Authorization`  | user JWT token | `Bearer ${token}` | ✓ |

### Poll from the queue

```
POST /pop
```
##### Request Body
```json
{
    "user": "bchadwic"
}
```
##### Response Body
```
Hello World!
```
#### Parameters
|     Field     |  Description  |  Constraints | Required |
| ------------- | ------------- | ------------ | ---- |
| `user`  | user to retreive queued text from |  `1-40` characters of `a-z`, `1-9` or `_` OR a valid email address | ✓ |

## FAQ

#### Why is the path to poll from a queue `/pop`?
Because you pop WordBubbles.


## Logging Standard
Log an `INFO` at the begining and end of each function, and be sure to include information that will help debug later on (ex: userId, trackingId, etc.)

Log an `ERROR` when an `error` is not `nil`. There is no need to log the `error` in multiple spots. Logging it at the source will help determine immediately where the problem is.

Log `DEBUG` and `WARN` sparsely. 

## Error Handling
For the time being, the strategy is to create the actual `error` response where the source of the `error` is. Creating the `error` message as deep as possible alleviates the `app` layer from being burdened with creating one. When the error comes from outside of this project, create one based off the function that threw the `error`.