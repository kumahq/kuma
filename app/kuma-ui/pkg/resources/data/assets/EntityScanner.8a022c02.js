import{_ as g,o as s,f as n,h as o,t as C,n as r,p as m,u as _,r as f,F as v,j as S,i as b,cP as k,d as h,w as y,e as w}from"./index.6180ff6f.js";const $={name:"FormFragment",props:{title:{type:String,required:!1,default:null},forAttr:{type:String,required:!1,default:null},allInline:{type:Boolean,default:!1},hideLabelCol:{type:Boolean,default:!1},equalCols:{type:Boolean,default:!1},shiftRight:{type:Boolean,default:!1}}},I={class:"form-line-wrapper"},z={key:0,class:"form-line__col"},E=["for"];function q(e,l,t,u,i,d){return s(),n("div",I,[o("div",{class:m(["form-line",{"has-equal-cols":t.equalCols}])},[t.hideLabelCol?r("",!0):(s(),n("div",z,[o("label",{for:t.forAttr,class:"k-input-label"},C(t.title)+": ",9,E)])),o("div",{class:m(["form-line__col",{"is-inline":t.allInline,"is-shifted-right":t.shiftRight}])},[_(e.$slots,"default")],2)],2)])}const re=g($,[["render",q],["__scopeId","data-v-b37aca73"]]);const x={props:{steps:{type:Array,default:()=>{}},sidebarContent:{type:Array,required:!0,default:()=>{}},footerEnabled:{type:Boolean,default:!0},nextDisabled:{type:Boolean,default:!0}},emits:["goToNextStep","goToStep","goToPrevStep"],data(){return{start:0}},computed:{step:{get(){return this.steps[this.start].slug},set(e){return this.steps[e].slug}},indexCanAdvance(){return this.start>=this.steps.length-1},indexCanReverse(){return this.start<=0}},watch:{"$route.query.step"(e=0){this.start!==e&&(this.start=e,this.$emit("goToNextStep",e))}},mounted(){this.resetProcess(),this.setStartingStep()},methods:{goToStep(e){this.start=e,this.updateQuery("step",e),this.$emit("goToStep",this.step)},goToNextStep(){this.start++,this.updateQuery("step",this.start),this.$emit("goToNextStep",this.step)},goToPrevStep(){this.start--,this.updateQuery("step",this.start),this.$emit("goToPrevStep",this.step)},updateQuery(e,l){const t=this.$router,u=this.$route;u.query?t.push({query:Object.assign({},u.query,{[e]:l})}):t.push({query:{[e]:l}})},setStartingStep(){const e=this.$route.query.step;this.start=e||0},resetProcess(){this.start=0,this.goToStep(0),localStorage.removeItem("storedFormData"),this.$refs.wizardForm.querySelectorAll('input[type="text"]').forEach(l=>{l.setAttribute("value","")})}}},F={class:"wizard-steps"},B={class:"wizard-steps__content-wrapper"},T={class:"wizard-steps__indicator"},R={class:"wizard-steps__indicator__controls",role:"tablist","aria-label":"steptabs"},N=["aria-selected","aria-controls"],A={class:"wizard-steps__content"},P={ref:"wizardForm",autocomplete:"off"},K=["id","aria-labelledby"],D={key:0,class:"wizard-steps__footer"},Q=w(" \u2039 Previous "),V=w(" Next \u203A "),L={class:"wizard-steps__sidebar"},j={class:"wizard-steps__sidebar__content"};function O(e,l,t,u,i,d){const p=f("KButton");return s(),n("div",F,[o("div",B,[o("header",T,[o("ul",R,[(s(!0),n(v,null,S(t.steps,(a,c)=>(s(),n("li",{key:a.slug,"aria-selected":d.step===a.slug?"true":"false","aria-controls":`wizard-steps__content__item--${c}`,class:m([{"is-complete":c<=i.start},"wizard-steps__indicator__item"])},[o("span",null,C(a.label),1)],10,N))),128))])]),o("div",A,[o("form",P,[(s(!0),n(v,null,S(t.steps,(a,c)=>(s(),n("div",{id:`wizard-steps__content__item--${c}`,key:a.slug,"aria-labelledby":`wizard-steps__content__item--${c}`,role:"tabpanel",tabindex:"0",class:"wizard-steps__content__item"},[d.step===a.slug?_(e.$slots,a.slug,{key:0},void 0,!0):r("",!0)],8,K))),128))],512)]),t.footerEnabled?(s(),n("footer",D,[b(h(p,{appearance:"outline",onClick:d.goToPrevStep},{default:y(()=>[Q]),_:1},8,["onClick"]),[[k,!d.indexCanReverse]]),b(h(p,{disabled:t.nextDisabled,appearance:"primary",onClick:d.goToNextStep},{default:y(()=>[V]),_:1},8,["disabled","onClick"]),[[k,!d.indexCanAdvance]])])):r("",!0)]),o("aside",L,[o("div",j,[(s(!0),n(v,null,S(t.sidebarContent,(a,c)=>(s(),n("div",{key:a.name,class:m(["wizard-steps__sidebar__item",`wizard-steps__sidebar__item--${c}`])},[_(e.$slots,a.name,{},void 0,!0)],2))),128))])])])}const ie=g(x,[["render",O],["__scopeId","data-v-83ee1f36"]]);const U={},G={class:"card-icon icon-success mb-3",role:"img"};function H(e,l,t,u,i,d){return s(),n("i",G," \u2713 ")}const J=g(U,[["render",H],["__scopeId","data-v-632fdab2"]]);const M={name:"EntityScanner",components:{IconSuccess:J},props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},shouldStart:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},emits:["hide-siblings"],data(){return{i:0,isRunning:!1,isComplete:!1,intervalId:null}},watch:{shouldStart(e,l){e!==l&&e===!0&&this.runScanner()}},mounted(){this.shouldStart===!0&&this.runScanner()},beforeUnmount(){clearInterval(this.intervalId)},methods:{runScanner(){this.isRunning=!0,this.isComplete=!1,this.intervalId=setInterval(()=>{this.i++,this.loaderFunction(),(this.i===this.retries||this.canComplete===!0)&&(clearInterval(this.intervalId),this.isRunning=!1,this.isComplete=!0,this.$emit("hide-siblings",!0))},this.interval)}}},W={key:0,class:"scanner"},X={class:"scanner-content"},Y={key:0,class:"card-icon mb-3"},Z={key:1,class:"card-icon mb-3"},ee={key:3},te={key:1};function se(e,l,t,u,i,d){const p=f("KIcon"),a=f("IconSuccess"),c=f("KEmptyState");return t.shouldStart?(s(),n("div",W,[o("div",X,[h(c,{"cta-is-hidden":""},{title:y(()=>[i.isRunning?(s(),n("div",Y,[h(p,{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"})])):r("",!0),i.isComplete&&t.hasError===!1&&i.isRunning===!1?(s(),n("div",Z,[h(a)])):r("",!0),i.isRunning?_(e.$slots,"loading-title",{key:2},void 0,!0):r("",!0),i.isRunning===!1?(s(),n("div",ee,[t.hasError?_(e.$slots,"error-title",{key:0},void 0,!0):r("",!0),i.isComplete&&t.hasError===!1?_(e.$slots,"complete-title",{key:1},void 0,!0):r("",!0)])):r("",!0)]),message:y(()=>[i.isRunning?_(e.$slots,"loading-content",{key:0},void 0,!0):r("",!0),i.isRunning===!1?(s(),n("div",te,[t.hasError?_(e.$slots,"error-content",{key:0},void 0,!0):r("",!0),i.isComplete&&t.hasError===!1?_(e.$slots,"complete-content",{key:1},void 0,!0):r("",!0)])):r("",!0)]),_:3})])])):r("",!0)}const ae=g(M,[["render",se],["__scopeId","data-v-5ffa385d"]]);export{ae as E,re as F,ie as S};
