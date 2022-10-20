import{cP as S,e as g,o,f as m,c as r,w as s,b as i,p,g as u,q as c,a as b,t as f,r as v}from"./index.f4381a04.js";const k={name:"OnboardingNavigation",props:{shouldAllowNext:{type:Boolean,default:!0},showSkip:{type:Boolean,default:!0},nextStep:{type:String,required:!0},previousStep:{type:String,default:""},nextStepTitle:{type:String,default:"Next"},lastStep:{type:Boolean,default:!1}},computed:{classes(){return["mt-4 flex items-center flex-col sm:flex-row",{"justify-center":this.lastStep,"justify-between":this.previousStep&&!this.lastStep,"justify-end":!this.previousStep&&!this.lastStep}]}},methods:{...S("onboarding",["completeOnboarding","changeStep"]),skipOnboarding(){this.completeOnboarding(),this.$router.push({name:"home"})}}};function _(l,t,e,x,h,a){const n=v("KButton");return o(),m("div",{class:c(a.classes)},[e.previousStep?(o(),r(n,{key:0,appearance:"primary",class:"navigation-button navigation-button--back",to:{name:e.previousStep},onClick:t[0]||(t[0]=d=>l.changeStep(e.previousStep))},{default:s(()=>[i(" Back ")]),_:1},8,["to"])):p("",!0),u("div",null,[e.showSkip?(o(),r(n,{key:0,class:"skip-button",appearance:"btn-link",size:"small",onClick:a.skipOnboarding},{default:s(()=>[i(" Skip Setup ")]),_:1},8,["onClick"])):p("",!0),u("span",{class:c(["inline-block",{"cursor-not-allowed":!e.shouldAllowNext}])},[b(n,{disabled:!e.shouldAllowNext,class:"navigation-button navigation-button--next",appearance:"primary",to:{name:e.nextStep},onClick:t[1]||(t[1]=d=>e.lastStep?a.skipOnboarding():l.changeStep(e.nextStep))},{default:s(()=>[i(f(e.nextStepTitle),1)]),_:1},8,["disabled","to"])],2)])],2)}const N=g(k,[["render",_],["__scopeId","data-v-134ba878"]]);export{N as O};
