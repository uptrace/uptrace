// decode query from url and transform into actual query

// updates labels from query

export function decodeQuery(query: string, apiUrl: string, labels = []): any {
  const queryArr = query
    ?.match(/[^{\}]+(?=})/g, '$1')
    ?.map((m) => m.split(','))
    ?.flat()

  const labelsFromQuery = []

  queryArr?.forEach((label) => {
    const regexQuery = label.match(/([^{}=,~!]+)/gm)
    if (!regexQuery) {
      return
    }
    if (label.includes('!=')) {
      const labelObj = {
        name: regexQuery[0],
        values: [],
      }
      const valueObj = {
        name: regexQuery[1]?.replaceAll('"', ''),
        selected: true,
        inverted: true,
      }

      labelObj.values.push(valueObj)
      labelsFromQuery.push(labelObj)
    } else if (label.includes('=~')) {
      const values = regexQuery[1]?.split('|')
      const labelObj = {
        name: regexQuery[0],
        values: [],
      }

      values.forEach((value) => {
        const valueObj = {
          name: value?.replaceAll('"', ''),
          selected: true,
          inverted: false,
        }

        labelObj.values.push(valueObj)
      })

      labelsFromQuery.push(labelObj)
    } else {
      const labelObj = {
        name: regexQuery[0],
        values: [],
      }
      const valueObj = {
        name: regexQuery[1]?.replaceAll('"', ''),
        selected: true,
        inverted: false,
      }
      labelObj.values.push(valueObj)
      labelsFromQuery.push(labelObj)
    }
  })

  const newLabels = [...labels] || []

  newLabels?.forEach((label) => {
    if (label.selected && label.values.length > 0) {
      label.selected = false
      label.values.forEach((value) => {
        if (value.selected) {
          value.selected = false
        }
      })
    }
  })

  if (labelsFromQuery.length > 0) {
    let labelsWithValues = []

    labelsFromQuery.forEach(async (label) => {
      const cleanLabel = newLabels?.find((item) => item?.name === label?.name)
      if (!cleanLabel) {
        return
      }

      await store.dispatch(loadLabelValues(cleanLabel, newLabels, apiUrl))

      // get actual labels selelcted
      const labelsWithValues = store.getState().labels
      const labelWithValues = labelsWithValues.find((item) => item?.name === label?.name)
      let values = labelWithValues?.values

      values = label?.values?.concat(values)

      values = values
        .sort((a, b) => a.name.localeCompare(b.name))
        .filter((value, index, arr) => {
          return value?.name !== arr?.[index - 1]?.name
        })
        .filter((value) => !!value)

      labelWithValues.values = values
      labelWithValues.selected = true
    })
    // set labels updated
    // store.dispatch(setLabels(labelsWithValues))
    return labelsWithValues
  }
}
