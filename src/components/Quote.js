import React, { useEffect } from 'react'
import { API } from 'aws-amplify'

export default function Quote() {
  const [quote, setQuote] = React.useState({author:"", quote:""})
  const [enableRefresh, setEnableRefresh] = React.useState(true)

  function updateQuote() {
    setEnableRefresh(false)

    API.get('QuotesAPI', '/quotes', {}).then(response => {
      console.log(response)
      setQuote(response);
    }).catch(error => {
      console.log(error.response);
    }).finally(() => {
      setEnableRefresh(true)
    })
  }

  useEffect(() => {
    updateQuote()
  }, [])

  if (quote.quote === "") {
    return null
  }

  return (
    <section>
    <div className="container-div">
      <div className='quote-div'>{quote.quote}</div>
      <div className='author-div'>- {quote.author}</div>
      <div className='button-div'><button onClick={updateQuote} disabled={!enableRefresh}>New Quote</button></div>
    </div>
    </section>
  )
}
