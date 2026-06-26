package onboarding

import "github.com/slack-go/slack"

const engineersPlaceholder = "Taro Yamamoto taro.yamamoto@global.komatsu\n(one engineer per line)"

func buildOnboardModal() slack.ModalViewRequest {
	return slack.ModalViewRequest{
		Type:       slack.VTModal,
		CallbackID: onboardModalCallback,
		Title:      plainText("Onboard engineers"),
		Submit:     plainText("File ticket"),
		Close:      plainText("Cancel"),
		Blocks: slack.Blocks{BlockSet: []slack.Block{
			multilineInput(blockEngineers, actionEngineers, "Engineers", engineersPlaceholder),
			staticSelect(blockLocation, actionLocation, "Office location", locations, defaultLocation),
			staticSelect(blockPriority, actionPriority, "Priority", priorities, defaultPriority),
			optionalInput(blockChannels, actionChannels, "Additional Slack channels (optional)", "#ext-channel1, #ext-channel2"),
		}},
	}
}

func buildSlackAddModal() slack.ModalViewRequest {
	return slack.ModalViewRequest{
		Type:       slack.VTModal,
		CallbackID: slackAddModalCallback,
		Title:      plainText("Add to Slack channels"),
		Submit:     plainText("File ticket"),
		Close:      plainText("Cancel"),
		Blocks: slack.Blocks{BlockSet: []slack.Block{
			multilineInput(blockEngineers, actionEngineers, "Engineers", engineersPlaceholder),
			textInputWithDefault(blockChannels, actionChannels, "Slack channels", "#channel1, #channel2", defaultSlackChannels),
		}},
	}
}

func plainText(s string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(slack.PlainTextType, s, false, false)
}

func optionalInput(blockID, actionID, label, placeholder string) *slack.InputBlock {
	element := slack.NewPlainTextInputBlockElement(plainText(placeholder), actionID)
	block := slack.NewInputBlock(blockID, plainText(label), nil, element)
	block.Optional = true
	return block
}

func multilineInput(blockID, actionID, label, placeholder string) *slack.InputBlock {
	element := slack.NewPlainTextInputBlockElement(plainText(placeholder), actionID)
	element.Multiline = true
	return slack.NewInputBlock(blockID, plainText(label), nil, element)
}

func textInputWithDefault(blockID, actionID, label, placeholder, initial string) *slack.InputBlock {
	element := slack.NewPlainTextInputBlockElement(plainText(placeholder), actionID)
	element.InitialValue = initial
	return slack.NewInputBlock(blockID, plainText(label), nil, element)
}

func staticSelect(blockID, actionID, label string, opts []orderedKV, defaultKey string) *slack.InputBlock {
	optionObjs := make([]*slack.OptionBlockObject, 0, len(opts))
	var initial *slack.OptionBlockObject
	for _, o := range opts {
		ob := slack.NewOptionBlockObject(o.Key, plainText(o.Key), nil)
		optionObjs = append(optionObjs, ob)
		if o.Key == defaultKey {
			initial = ob
		}
	}
	element := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, plainText("Select..."), actionID, optionObjs...)
	if initial != nil {
		element.InitialOption = initial
	}
	return slack.NewInputBlock(blockID, plainText(label), nil, element)
}
