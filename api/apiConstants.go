package api

const twentyfourHours = 60 * 60 * 24
const twelveHours = 60 * 60 * 12
const oneEightyDays = 60 * 60 * 24 * 180

const maxImageDimension = 10000
const minImageDimension = 400
const thumbnailsSize = 150
const resizeSize = 3840

const emailRegex = `^[^\s@]+@[^\s@]+\.[^\s@]+$`
const nameRegex = `^([a-z]|[0-9]|-)+$`

const sessionTokenKey = "session_token"
const jpegExtension = ".jpg"

const filesDirectory = "/fourame"
const userDirectory = "/users"
const thumbnailsDirectory = "/thumbnails"
const imagesDiretory = "/images"
