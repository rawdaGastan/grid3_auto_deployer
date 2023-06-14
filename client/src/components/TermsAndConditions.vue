<template>
	<v-dialog v-model="dialog" transition="dialog-top-transition" width="800">
		<template v-slot:activator="{ props }">
			<v-row class="ms-2 my-2">
				<v-checkbox :checked="modelValue" v-model="checked" :rules="termsPrivacyError" density="compact"
					@change="checked = terms || checked; $event.target.checked = checked; terms = false;">
					<template v-slot:label>
						<a class="grey-darken-4" v-bind="props" @click="terms = true;">
							Terms and Conditions
						</a>
					</template>
				</v-checkbox>
			</v-row>
		</template>
		<v-card class="pa-5">
			<v-card-title class="text-center">
				<span class="text-h5 primary">Codescalers Students terms and conditions
				</span>
			</v-card-title>
			<v-card-text>
				<p>
					Codescalers Egypt (the company) is offering its cloud resources to
					university senior students (user/users) across Egypt for their
					graduation project. User/Users will comply with the terms of this
					Agreement and the <strong>Privacy Policy</strong> and any other terms
					and conditions required by in connection herewith, the Open Source
					Licenses, and the terms of all other agreements to which User/Users
					are a party in connection with User/Users performance under this
					Agreement including, without limitation, any agreement User/Users have
					with a third-party service provider.
				</p>
				<v-list lines="two">
					<v-list-item>
						User/users will use and operate the resources in strict compliance
						with terms of this Agreement and any applicable laws or regulations,
						and User/Users will not take any action not expressly authorized
						hereunder.
					</v-list-item>
					<v-list-item>
						Without prejudice to User/Users rights under any applicable Open
						Source license, User/Users will not modify or attempt to modify the
						resources for any purpose including but not limited to attempting to
						circumvent the audit, bypass security, manipulate the performance
						of, or otherwise disrupt the resources for any reason, including but
						not limited to attempting to increase the amount of data stored or
						bandwidth utilized, as defined herein, and User/Users will not
						otherwise interfere with the operation of the resources.
					</v-list-item>
					<v-list-item>
						User/Users acknowledge and agree that the Company has no practical
						access to User/Users data or knowledge of the nature of the data
						stored on the grid (including but not limited to Content) and that
						it does not retain any Content or other data that User/Users process
						using the company resources. The Company has no control in the
						management of such Content or data, nor any influence in the
						specific processing procedures. The Company will never pursue
						changes to the resources that could make User data accessible to
						Company or third parties.
					</v-list-item>
				</v-list>
				<h5 class="text-h5 my-5 font-weight-bold">CONTENT RESTRICTIONS</h5>
				<p>
					The IT Capacity that is made available can be used by any User/Users
					to create, post, upload, share or store Content, such as text,
					graphics, photos, videos, sound, data or other information and
					materials submitted or provided by Users.
				</p>
				<p>
					The company has no access to, nor do we control, own, or endorse any
					Content that User/Users transmit, store or process. User/Users are
					solely responsible for any Content stored using the resources, and for
					any data that User/Users have entered.
				</p>
				<p>
					User/Users hereby represent and warrant that: (1) User/Users own all
					intellectual property rights (or have obtained all necessary
					permissions) to provide User/Users Content and other data; (2) such
					Content and other data will not violate any agreements or
					confidentiality obligations; and (3) such Content and data will not
					violate, infringe or misappropriate any intellectual property right or
					other proprietary rights of any person or third party.
				</p>
				<p>
					User/Users will not create, post, share or store Content or other data
					that:
				</p>
				<v-list lines="one">
					<v-list-item>
						is unlawful, libelous, defamatory, harassing, threatening, invasive
						of privacy or publicity rights, abusive, inflammatory, fraudulent or
						otherwise objectionable;
					</v-list-item>
					<v-list-item>
						would constitute, encourage or provide instructions for a criminal
						offense, violate the rights of any party, otherwise create liability
						or violate any local, state, national or international law;
					</v-list-item>
					<v-list-item>
						intentionally misleads by containing or depicting any statements,
						remarks or claims that do not reflect User/Users honest views and
						experiences;
					</v-list-item>
					<v-list-item>
						impersonates, or misrepresents User/Users affiliation with, any
						person or entity (including Company);
					</v-list-item>
					<v-list-item>
						references or depicts Company, our TF Products or any related
						services but fails to disclose any material connection to us that
						may exist;
					</v-list-item>
					<v-list-item>
						contains any unsolicited promotions, political campaigning,
						advertising or solicitations;
					</v-list-item>
					<v-list-item>
						contains any viruses, corrupted data or other harmful, disruptive or
						destructive files or content; or
					</v-list-item>
					<v-list-item>
						in our sole judgment, is objectionable or that restricts or inhibits
						any other person from using or enjoying the TF Products, or that may
						expose Company or others to any harm or liability of any type;
					</v-list-item>
				</v-list>
				<h5 class="text-h5 my-5 font-weight-bold">
					USAGE RESTRICTIONS (ACCEPTABLE USE)
				</h5>
				<p>
					User/Users will use the resources in strict accordance with the terms
					of this Agreement and in no other manner. Without limiting the
					generality of the foregoing, User/Users will not:
				</p>
				<v-list lines="two">
					<v-list-item>
						access or use the resources: (i) in violation of applicable laws;
						(ii) to send or store material knowingly or intentionally containing
						software viruses, worms, Trojan horses or other harmful computer
						code, files, or scripts; (iii) in a manner that interferes with or
						disrupts the integrity or performance of the Grid (or the data
						contained therein); or (iv) in any manner that could interfere with,
						disrupt, negatively affect or inhibit other Users from fully
						enjoying their use of the resources or that could damage, disable,
						overburden or impair the functioning of the resources in any manner.
					</v-list-item>
					<v-list-item>
						delete or otherwise render Content unavailable for recovery
						independent of the programmatic functionality of the relevant
						software;
					</v-list-item>
					<v-list-item>
						violate, infringe or misappropriate any intellectual property or
						other third-party right or commit a tort; experiences;
					</v-list-item>
					<v-list-item>
						modify, copy (other than standard page caching), publicly perform,
						publicly display, sell, rent, lease, timeshare or otherwise
						distribute the resources, in whole or in part. This restriction does
						not apply to open source software we release, which User/Users can
						use subject to the applicable open source software license terms;
					</v-list-item>
					<v-list-item>
						attempt to bypass or circumvent measures employed to prevent or
						limit access to any content, area or functionality on the resources,
						without providing prior notice to Company of the method used to
						bypass or circumvent;
					</v-list-item>
					<v-list-item>
						in any other way attempt to interfere, impede, alter, or otherwise
						interact with the resources in any manner not expressly authorized
						hereunder;
					</v-list-item>
					<v-list-item>
						use any of the resources other than for its intended purposes; or
					</v-list-item>
					<v-list-item>
						use the resources to engage in or promote any activity that violates
						these User Terms and Conditions.
					</v-list-item>

					<v-list-item>
						Any violation of one or more of the terms and conditions of this
						agreement, gives automatically the rights to the company to
						terminate the User/Users resources without notifying them.
					</v-list-item>

					<v-list-item>
						Company has the right to suspend or end the User/Users use of the
						resources if the Company believes the User/Users are using the
						resources for a purpose other than the graduation project.
					</v-list-item>

					<v-list-item>
						Company has the right to suspend or end the User's use of the
						resources if Company believes the User to be: (a) violating any term
						of this Agreement; or (b) using the resources in a manner that
						Company reasonably believes may cause a security risk, a disruption
						to the company, or liability for Company or any persons involved in
						the company Open Source project.
					</v-list-item>

					<v-list-item>
						The Company has the right to end/terminate the resources of all
						users by 15th of September 2023
					</v-list-item>
				</v-list>
			</v-card-text>
			<v-card-actions>
				<v-spacer></v-spacer>
				<v-btn color="green-darken-1" variant="text" @click="checked = true; dialog = false;">
					Accept
				</v-btn>
				<v-btn color="red-darken-1" variant="text" @click="checked = false; dialog = false;">
					Decline
				</v-btn>
			</v-card-actions>
		</v-card>
	</v-dialog>
</template>

<script>
import { ref } from "vue";
export default {
	emits: ["update:modelValue"],
	props: {
		modelValue: {
			type: Boolean,
			required: true,
		},
	},
	data() {
		return {
			checked: this.modelValue
		}
	},
	setup() {
		const dialog = ref(false);
		const terms = ref(false);
		const termsPrivacyError = ref([
			(value) => {
				if (!value) return "You have to agree the terms and privacy condition";
				return true;
			},
		]);

		return { dialog, terms, termsPrivacyError };
	},
};
</script>
