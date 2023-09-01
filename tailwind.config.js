/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./cmd/frogboard/templates/*.tmpl.html", "./cmd/frogboard/templates/required/*.tmpl.html", "./internal/models/post.go"],
  theme: {
    extend: {},
  },
  plugins: [],
}

