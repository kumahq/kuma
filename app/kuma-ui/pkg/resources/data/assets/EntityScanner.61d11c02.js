import{_ as g,o as s,e as n,f as o,t as w,k as r,n as m,A as u,cy as $,F as v,h as S,g as b,cP as k,a as p,w as y,b as C,r as f}from"./index.dbfc69fe.js";const z={name:"FormFragment",props:{title:{type:String,required:!1,default:null},forAttr:{type:String,required:!1,default:null},allInline:{type:Boolean,default:!1},hideLabelCol:{type:Boolean,default:!1},equalCols:{type:Boolean,default:!1},shiftRight:{type:Boolean,default:!1}}},I={class:"form-line-wrapper"},E={key:0,class:"form-line__col"},q=["for"];function x(e,l,t,_,i,c){return s(),n("div",I,[o("div",{class:m(["form-line",{"has-equal-cols":t.equalCols}])},[t.hideLabelCol?r("",!0):(s(),n("div",E,[o("label",{for:t.forAttr,class:"k-input-label"},w(t.title)+": ",9,q)])),o("div",{class:m(["form-line__col",{"is-inline":t.allInline,"is-shifted-right":t.shiftRight}])},[u(e.$slots,"default")],2)],2)])}const ne=g(z,[["render",x],["__scopeId","data-v-b37aca73"]]);const F={props:{steps:{type:Array,default:()=>{}},sidebarContent:{type:Array,required:!0,default:()=>{}},footerEnabled:{type:Boolean,default:!0},nextDisabled:{type:Boolean,default:!0}},emits:["goToNextStep","goToStep","goToPrevStep"],data(){return{start:0}},computed:{step:{get(){return this.steps[this.start].slug},set(e){return this.steps[e].slug}},indexCanAdvance(){return this.start>=this.steps.length-1},indexCanReverse(){return this.start<=0}},watch:{"$route.query.step"(e=0){this.start!==e&&(this.start=e,this.$emit("goToNextStep",e))}},mounted(){this.resetProcess(),this.setStartingStep()},methods:{goToStep(e){this.start=e,this.updateQuery("step",e),this.$emit("goToStep",this.step)},goToNextStep(){this.start++,this.updateQuery("step",this.start),this.$emit("goToNextStep",this.step)},goToPrevStep(){this.start--,this.updateQuery("step",this.start),this.$emit("goToPrevStep",this.step)},updateQuery(e,l){const t=this.$router,_=this.$route;_.query?t.push({query:Object.assign({},_.query,{[e]:l})}):t.push({query:{[e]:l}})},setStartingStep(){const e=this.$route.query.step;this.start=e||0},resetProcess(){this.start=0,this.goToStep(0),$.remove("storedFormData"),this.$refs.wizardForm.querySelectorAll('input[type="text"]').forEach(l=>{l.setAttribute("value","")})}}},B={class:"wizard-steps"},T={class:"wizard-steps__content-wrapper"},R={class:"wizard-steps__indicator"},N={class:"wizard-steps__indicator__controls",role:"tablist","aria-label":"steptabs"},A=["aria-selected","aria-controls"],P={class:"wizard-steps__content"},K={ref:"wizardForm",autocomplete:"off"},D=["id","aria-labelledby"],Q={key:0,class:"wizard-steps__footer"},V={class:"wizard-steps__sidebar"},L={class:"wizard-steps__sidebar__content"};function j(e,l,t,_,i,c){const h=f("KButton");return s(),n("div",B,[o("div",T,[o("header",R,[o("ul",N,[(s(!0),n(v,null,S(t.steps,(a,d)=>(s(),n("li",{key:a.slug,"aria-selected":c.step===a.slug?"true":"false","aria-controls":`wizard-steps__content__item--${d}`,class:m([{"is-complete":d<=i.start},"wizard-steps__indicator__item"])},[o("span",null,w(a.label),1)],10,A))),128))])]),o("div",P,[o("form",K,[(s(!0),n(v,null,S(t.steps,(a,d)=>(s(),n("div",{id:`wizard-steps__content__item--${d}`,key:a.slug,"aria-labelledby":`wizard-steps__content__item--${d}`,role:"tabpanel",tabindex:"0",class:"wizard-steps__content__item"},[c.step===a.slug?u(e.$slots,a.slug,{key:0},void 0,!0):r("",!0)],8,D))),128))],512)]),t.footerEnabled?(s(),n("footer",Q,[b(p(h,{appearance:"outline",onClick:c.goToPrevStep},{default:y(()=>[C(" \u2039 Previous ")]),_:1},8,["onClick"]),[[k,!c.indexCanReverse]]),b(p(h,{disabled:t.nextDisabled,appearance:"primary",onClick:c.goToNextStep},{default:y(()=>[C(" Next \u203A ")]),_:1},8,["disabled","onClick"]),[[k,!c.indexCanAdvance]])])):r("",!0)]),o("aside",V,[o("div",L,[(s(!0),n(v,null,S(t.sidebarContent,(a,d)=>(s(),n("div",{key:a.name,class:m(["wizard-steps__sidebar__item",`wizard-steps__sidebar__item--${d}`])},[u(e.$slots,a.name,{},void 0,!0)],2))),128))])])])}const re=g(F,[["render",j],["__scopeId","data-v-269f0219"]]);const O={},U={class:"card-icon icon-success mb-3",role:"img"};function G(e,l,t,_,i,c){return s(),n("i",U," \u2713 ")}const H=g(O,[["render",G],["__scopeId","data-v-632fdab2"]]);const J={name:"EntityScanner",components:{IconSuccess:H},props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},shouldStart:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},emits:["hide-siblings"],data(){return{i:0,isRunning:!1,isComplete:!1,intervalId:null}},watch:{shouldStart(e,l){e!==l&&e===!0&&this.runScanner()}},mounted(){this.shouldStart===!0&&this.runScanner()},beforeUnmount(){clearInterval(this.intervalId)},methods:{runScanner(){this.isRunning=!0,this.isComplete=!1,this.intervalId=setInterval(()=>{this.i++,this.loaderFunction(),(this.i===this.retries||this.canComplete===!0)&&(clearInterval(this.intervalId),this.isRunning=!1,this.isComplete=!0,this.$emit("hide-siblings",!0))},this.interval)}}},M={key:0,class:"scanner"},W={class:"scanner-content"},X={key:0,class:"card-icon mb-3"},Y={key:1,class:"card-icon mb-3"},Z={key:3},ee={key:1};function te(e,l,t,_,i,c){const h=f("KIcon"),a=f("IconSuccess"),d=f("KEmptyState");return t.shouldStart?(s(),n("div",M,[o("div",W,[p(d,{"cta-is-hidden":""},{title:y(()=>[i.isRunning?(s(),n("div",X,[p(h,{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"})])):r("",!0),i.isComplete&&t.hasError===!1&&i.isRunning===!1?(s(),n("div",Y,[p(a)])):r("",!0),i.isRunning?u(e.$slots,"loading-title",{key:2},void 0,!0):r("",!0),i.isRunning===!1?(s(),n("div",Z,[t.hasError?u(e.$slots,"error-title",{key:0},void 0,!0):r("",!0),i.isComplete&&t.hasError===!1?u(e.$slots,"complete-title",{key:1},void 0,!0):r("",!0)])):r("",!0)]),message:y(()=>[i.isRunning?u(e.$slots,"loading-content",{key:0},void 0,!0):r("",!0),i.isRunning===!1?(s(),n("div",ee,[t.hasError?u(e.$slots,"error-content",{key:0},void 0,!0):r("",!0),i.isComplete&&t.hasError===!1?u(e.$slots,"complete-content",{key:1},void 0,!0):r("",!0)])):r("",!0)]),_:3})])])):r("",!0)}const ie=g(J,[["render",te],["__scopeId","data-v-5ffa385d"]]);export{ie as E,ne as F,re as S};
