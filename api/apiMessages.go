package api

const errLogTemplate = "EVENT: %q SERVICE: %q, REQUEST: %q REASON: %q"
const errLogCannotDecode = "Cannot decode the body of request."
const errLogValidation = "Invalid request."
const errLogCacheFailure = "Accessing cache error."
const errLogAlreadyExists = "Already exists."
const errLogEmailFailure = "Sending Email Failed."
const errLogTooManyTries = "Too may tries."
const errLogNotFound = "Not found."
const errLogWrongVerificationCode = "Wrong verification code."
const errLogInitialRequestLost = "The initial request got lost."
const errLogCannotConnectToDb = "Cannot connect to DB."
const errLogCannotInsertToDb = "Cannot insert to DB."
const errLogCannotUpdateTheDb = "Cannot update DB."
const errLogCannotRetrieveFromDb = "Cannot retrieve from DB."
const errLogDb = "DB error."
const errLogWrongCredentials = "Wrong credentials."
const errLogIvalidSessionToken = "Invalid Cookie"
const errLogQuotaExceeded = "Quota Exceeded."
const errLogImageUploadError = "Image Upload Error."
const errLogImageValidationError = "Image Validation Error."
const errLogImageSavingError = "Image Saving Error."
const errLogMissingField = "Missing Field."
const errLogIoError = "IO Error."

const errCannotDecode = "Invalid JSON object."
const errInternalError = "Processing request failed because of an internal error"
const errInvalidEmailFormat = "email format is not valid"
const errTooMayTries = "Only three tries is allowed every 12 hours."
const errRegistrationNotFound = "The account is not pending for verification or the registration has timed out."
const errWrongVerificationCode = "The verification code is wrong."
const errFailedLogin = "Email or password is wrong."
const errUnsupportedOperation = "Unsupported operation."
const errUnAuthorized = "Unauthorized."
const errBadRequest = "Bad request."
const errQuotaExceeded = "Quota Exceeded."
const errUnsupportedImage = "Uploaded Image is not supported. Currently, only jpeg images are supported."
const errImageTooBig = "Image is too big the max dimension supported is 10,000 pixel."
const errImageTooSmall = "Image is too small the min dimension supported is 400 pixel."
const errNotFound = "Image with such id was not found."

const emailSubject = "Verification Code"
const emailBody = "Your verification code is: "
