import type {
	AdminSemesterCalendar,
	AdminSemesterCalendarUpsertRequest,
	CustomWeekRange
} from '$lib/admin-api';

export type NoteRowForm = {
	row: string;
	content: string;
	needNumber: boolean;
};

export type CustomWeekRangeForm = {
	startRow: string;
	endRow: string;
	content: string;
};

export type SemesterCalendarFormState = {
	semesterCode: string;
	title: string;
	subtitle: string;
	calendarStart: string;
	calendarEnd: string;
	semesterStart: string;
	semesterEnd: string;
	notes: NoteRowForm[];
	customWeekRanges: CustomWeekRangeForm[];
};

export function createEmptyNote(): NoteRowForm {
	return {
		row: '',
		content: '',
		needNumber: false
	};
}

export function createEmptyCustomWeekRange(): CustomWeekRangeForm {
	return {
		startRow: '',
		endRow: '',
		content: ''
	};
}

export function createEmptySemesterCalendarForm(): SemesterCalendarFormState {
	return {
		semesterCode: '',
		title: '',
		subtitle: '',
		calendarStart: '',
		calendarEnd: '',
		semesterStart: '',
		semesterEnd: '',
		notes: [createEmptyNote()],
		customWeekRanges: [createEmptyCustomWeekRange()]
	};
}

export function toApiDate(value: string): string {
	return `${value}T00:00:00Z`;
}

export function fromApiDate(value: string): string {
	return value.slice(0, 10);
}

export function fromAdminSemesterCalendar(item: AdminSemesterCalendar): SemesterCalendarFormState {
	return {
		semesterCode: item.semesterCode,
		title: item.title,
		subtitle: item.subtitle,
		calendarStart: fromApiDate(item.calendarStart),
		calendarEnd: fromApiDate(item.calendarEnd),
		semesterStart: fromApiDate(item.semesterStart),
		semesterEnd: fromApiDate(item.semesterEnd),
		notes:
			item.notes.length > 0
				? item.notes.map((note) => ({
						row: String(note.row),
						content: note.content,
						needNumber: Boolean(note.needNumber)
					}))
				: [createEmptyNote()],
		customWeekRanges:
			item.customWeekRanges.length > 0
				? item.customWeekRanges.map((range) => ({
						startRow: String(range.startRow),
						endRow: String(range.endRow),
						content: range.content
					}))
				: [createEmptyCustomWeekRange()]
	};
}

export function buildSemesterCalendarPayload(
	form: SemesterCalendarFormState
): { payload: AdminSemesterCalendarUpsertRequest; error: '' } | { payload: null; error: string } {
	const semesterCode = form.semesterCode.trim();
	const title = form.title.trim();
	const subtitle = form.subtitle.trim();

	if (
		!semesterCode ||
		!title ||
		!subtitle ||
		!form.calendarStart ||
		!form.calendarEnd ||
		!form.semesterStart ||
		!form.semesterEnd
	) {
		return { payload: null, error: '请填写完整内容' };
	}

	if (form.calendarStart > form.calendarEnd) {
		return { payload: null, error: '校历开始日期不能晚于校历结束日期' };
	}

	if (form.semesterStart > form.semesterEnd) {
		return { payload: null, error: '学期开始日期不能晚于学期结束日期' };
	}

	const notes = [];
	for (const item of form.notes) {
		const row = item.row.trim();
		const content = item.content.trim();
		if (!row && !content) {
			continue;
		}

		const rowValue = Number(row);
		if (!row || !Number.isInteger(rowValue) || rowValue <= 0 || !content) {
			return { payload: null, error: '请填写有效的备注信息' };
		}

		notes.push({
			row: rowValue,
			content,
			needNumber: item.needNumber
		});
	}

	const customWeekRanges: CustomWeekRange[] = [];
	for (const item of form.customWeekRanges) {
		const startRow = item.startRow.trim();
		const endRow = item.endRow.trim();
		const content = item.content.trim();
		if (!startRow && !endRow && !content) {
			continue;
		}

		const startRowValue = Number(startRow);
		const endRowValue = Number(endRow);
		if (
			!startRow ||
			!endRow ||
			!content ||
			!Number.isInteger(startRowValue) ||
			!Number.isInteger(endRowValue) ||
			startRowValue <= 0 ||
			endRowValue <= 0
		) {
			return { payload: null, error: '请填写有效的自定义周次范围' };
		}

		if (startRowValue > endRowValue) {
			return { payload: null, error: '自定义周次范围的开始行不能大于结束行' };
		}

		customWeekRanges.push({
			startRow: startRowValue,
			endRow: endRowValue,
			content
		});
	}

	return {
		payload: {
			semesterCode,
			title,
			subtitle,
			calendarStart: toApiDate(form.calendarStart),
			calendarEnd: toApiDate(form.calendarEnd),
			semesterStart: toApiDate(form.semesterStart),
			semesterEnd: toApiDate(form.semesterEnd),
			notes,
			customWeekRanges
		},
		error: ''
	};
}
