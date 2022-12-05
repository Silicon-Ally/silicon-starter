// This is a test example of how to use Jest in unit tests.
// It can be deleted once other unit tests exist.

export const colatzCount = (a: number): number => {
  if (!Number.isInteger(a)) {
    throw `The colatz conjecture is only well defined for integers, was given a real number ${a}`
  }
  if (a <= 0) {
    throw `The colatz conjecture is only well defined for positive integers , was given a negative number ${a}`
  }
  let count = 0
  while (a !== 1) {
    count++
    if (a % 2 === 0) {
      a = a / 2
    } else {
      a = a * 3 + 1
    }
  }
  return count
}