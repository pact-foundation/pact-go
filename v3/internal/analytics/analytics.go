package analytics

// TEST
// process.env.GA_ACCOUNT_ID = "UA-8926693-14"
// export const sendToGA = ({category, action, label, value, user = "", tenant = "", client = "", clientLibraryAgent = "", plan = ""}) => {
//   // Event Format: https://developers.google.com/analytics/devguides/collection/protocol/v1/devguide#batch
//   const events = `v=1&ds=api&tid=${GA_ACCOUNT_ID}&cid=${uuid.v4()}&t=event&ec=${category}&ea=${action}&el=${label}&ev=${value}&cd1=${tenant}&cd5=${user}&cd6=${client}&cd7=${clientLibraryAgent}&cd2=${plan}`

//   return axios
//     .post(GA_URL, events, {
//       headers: DEFAULT_HEADERS,
//     })
//     .catch(err => {
//       log.error('error sending to ga', {}, err)
//     })
// }

// TODO:

// 1. Record download: OS, Golang version, Lib version
// 2. Run HTTP unit test
// 3.
// 3. Verify
