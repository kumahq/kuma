import{k as b,O as g,cN as S,cc as m,o as s,c as k,i as r,w as i,b as t,j as d,e as u,ci as p,a as v,bV as f}from"./index.0a811bc4.js";const x={name:"OnboardingNavigation",components:{KButton:g},props:{shouldAllowNext:{type:Boolean,default:!0},showSkip:{type:Boolean,default:!0},nextStep:{type:String,required:!0},previousStep:{type:String,default:""},nextStepTitle:{type:String,default:"Next"},lastStep:{type:Boolean,default:!1}},computed:{classes(){return["mt-4 flex items-center flex-col sm:flex-row",{"justify-center":this.lastStep,"justify-between":this.previousStep&&!this.lastStep,"justify-end":!this.previousStep&&!this.lastStep}]}},methods:{...S("onboarding",["completeOnboarding","changeStep"]),skipOnboarding(){this.completeOnboarding(),this.$router.push({name:"home"})}}};function _(l,a,e,h,y,n){const o=m("KButton");return s(),k("div",{class:p(n.classes)},[e.previousStep?(s(),r(o,{key:0,appearance:"primary",class:"navigation-button navigation-button--back",to:{name:e.previousStep},"data-testid":"onboarding-previous-button",onClick:a[0]||(a[0]=c=>l.changeStep(e.previousStep))},{default:i(()=>[t(`
      Back
    `)]),_:1},8,["to"])):d("",!0),t(),u("div",null,[e.showSkip?(s(),r(o,{key:0,class:"skip-button",appearance:"btn-link",size:"small","data-testid":"onboarding-skip-button",onClick:n.skipOnboarding},{default:i(()=>[t(`
        Skip Setup
      `)]),_:1},8,["onClick"])):d("",!0),t(),u("span",{class:p(["inline-block",{"cursor-not-allowed":!e.shouldAllowNext}])},[v(o,{disabled:!e.shouldAllowNext,class:"navigation-button navigation-button--next",appearance:"primary",to:{name:e.nextStep},"data-testid":"onboarding-next-button",onClick:a[1]||(a[1]=c=>e.lastStep?n.skipOnboarding():l.changeStep(e.nextStep))},{default:i(()=>[t(f(e.nextStepTitle),1)]),_:1},8,["disabled","to"])],2)])],2)}const B=b(x,[["render",_],["__scopeId","data-v-779a1d51"]]);export{B as O};
