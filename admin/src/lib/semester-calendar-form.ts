import type {
	AdminSemesterCalendar,
	AdminSemesterCalendarUpsertRequest,
	CustomWeekRange
} from '$lib/admin-api';

export type NoteRowForm = {
	row: number;
	content: string;
	needNumber: boolean;
};

export type CustomWeekRangeForm = {
	startRow: number;
	endRow: number;
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
		row: NaN,
		content: '',
		needNumber: false
	};
}

export function createEmptyCustomWeekRange(): CustomWeekRangeForm {
	return {
		startRow: NaN,
		endRow: NaN,
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
						row: note.row,
						content: note.content,
						needNumber: Boolean(note.needNumber)
					}))
				: [createEmptyNote()],
		customWeekRanges:
			item.customWeekRanges.length > 0
				? item.customWeekRanges.map((range) => ({
						startRow: range.startRow,
						endRow: range.endRow,
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
		const row = item.row;
		const content = item.content.trim();
		if (isNaN(row) && !content) {
			continue;
		}

		if (isNaN(row) || !Number.isInteger(row) || row <= 0 || !content) {
			return { payload: null, error: '请填写有效的备注信息' };
		}

		notes.push({
			row,
			content,
			needNumber: item.needNumber
		});
	}

	const customWeekRanges: CustomWeekRange[] = [];
	for (const item of form.customWeekRanges) {
		const startRow = item.startRow;
		const endRow = item.endRow;
		const content = item.content.trim();
		if (isNaN(startRow) && isNaN(endRow) && !content) {
			continue;
		}

		if (
			isNaN(startRow) ||
			isNaN(endRow) ||
			!Number.isInteger(startRow) ||
			!Number.isInteger(endRow) ||
			startRow <= 0 ||
			endRow <= 0 ||
			!content
		) {
			return { payload: null, error: '请填写有效的自定义周次范围' };
		}

		if (startRow > endRow) {
			return { payload: null, error: '自定义周次范围的开始行不能大于结束行' };
		}

		customWeekRanges.push({
			startRow,
			endRow,
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
