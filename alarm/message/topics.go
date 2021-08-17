package message


/*
	JSON example payload:
		{
			AlarmID: "e36a6c22-ece6-46eb-9016-9303273edbfe",
			UserID: "e859dab9-66d4-4bf4-a578-113d223b94f0",
			Status: "WARNING",
			ChangedAt: "2021-06-07T20:40:15.598765212Z"
		}
*/
const AlarmStatusChangedTopic = "AlarmStatusChanged"




/*
	JSON example payload:
		{
			UserID: "e859dab9-66d4-4bf4-a578-113d223b94f0",
		}
*/
const SendAlarmDigestTopic = "SendAlarmDigest"



/*
	JSON example payload:

		{
			UserID: "e859dab9-66d4-4bf4-a578-113d223b94f0",
			ActiveAlarms: [{
					AlarmID: "e36a6c22-ece6-46eb-9016-9303273edbfe",
					Status: "WARNING",
					LatestChangedAt: "2021-06-07T20:40:15.598765212Z"
				}, {
					AlarmID: "29f717c1-70b8-4f42-b2f5-89bf21f560e9",
					Status: "WARNING",
					LatestChangedAt: "2021-06-07T21:23:01.828161292Z"
			}]
		}
*/
const AlarmDigestTopic = "AlarmDigest"
