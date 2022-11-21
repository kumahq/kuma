import{D as g,o as s,j as n,l as o,t as w,z as a,A as m,I as d,e as I,i as f,F as v,n as S,m as b,cV as k,a as p,w as y,b as C,O as z,K as E}from"./index.c8163df9.js";const x={name:"FormFragment",props:{title:{type:String,required:!1,default:null},forAttr:{type:String,required:!1,default:null},allInline:{type:Boolean,default:!1},hideLabelCol:{type:Boolean,default:!1},equalCols:{type:Boolean,default:!1},shiftRight:{type:Boolean,default:!1}}},B={class:"form-line-wrapper"},q={key:0,class:"form-line__col"},$=["for"];function F(e,c,t,u,r,_){return s(),n("div",B,[o("div",{class:m(["form-line",{"has-equal-cols":t.equalCols}])},[t.hideLabelCol?a("",!0):(s(),n("div",q,[o("label",{for:t.forAttr,class:"k-input-label"},w(t.title)+": ",9,$)])),o("div",{class:m(["form-line__col",{"is-inline":t.allInline,"is-shifted-right":t.shiftRight}])},[d(e.$slots,"default")],2)],2)])}const re=g(x,[["render",F],["__scopeId","data-v-62a81d56"]]);const R={components:{KButton:I},props:{steps:{type:Array,default:()=>{}},sidebarContent:{type:Array,required:!0,default:()=>{}},footerEnabled:{type:Boolean,default:!0},nextDisabled:{type:Boolean,default:!0}},emits:["goToStep"],data(){return{start:0}},computed:{step:{get(){return this.steps[this.start].slug},set(e){return this.steps[e].slug}},indexCanAdvance(){return this.start>=this.steps.length-1},indexCanReverse(){return this.start<=0}},mounted(){this.setStartingStep()},methods:{goToNextStep(){this.start++,this.updateQuery("step",this.start),this.$emit("goToStep",this.step)},goToPrevStep(){this.start--,this.updateQuery("step",this.start),this.$emit("goToStep",this.step)},updateQuery(e,c){const t=this.$router,u=this.$route;u.query?t.push({query:Object.assign({},u.query,{[e]:c})}):t.push({query:{[e]:c}})},setStartingStep(){const e=this.$route.query.step;this.start=e||0}}},K={class:"wizard-steps"},N={class:"wizard-steps__content-wrapper"},T={class:"wizard-steps__indicator"},A={class:"wizard-steps__indicator__controls",role:"tablist","aria-label":"steptabs"},D=["aria-selected","aria-controls"],V={class:"wizard-steps__content"},L={ref:"wizardForm",autocomplete:"off"},P=["id","aria-labelledby"],Q={key:0,class:"wizard-steps__footer"},j={class:"wizard-steps__sidebar"},O={class:"wizard-steps__sidebar__content"};function U(e,c,t,u,r,_){const h=f("KButton");return s(),n("div",K,[o("div",N,[o("header",T,[o("ul",A,[(s(!0),n(v,null,S(t.steps,(i,l)=>(s(),n("li",{key:i.slug,"aria-selected":_.step===i.slug?"true":"false","aria-controls":`wizard-steps__content__item--${l}`,class:m([{"is-complete":l<=r.start},"wizard-steps__indicator__item"])},[o("span",null,w(i.label),1)],10,D))),128))])]),o("div",V,[o("form",L,[(s(!0),n(v,null,S(t.steps,(i,l)=>(s(),n("div",{id:`wizard-steps__content__item--${l}`,key:i.slug,"aria-labelledby":`wizard-steps__content__item--${l}`,role:"tabpanel",tabindex:"0",class:"wizard-steps__content__item"},[_.step===i.slug?d(e.$slots,i.slug,{key:0},void 0,!0):a("",!0)],8,P))),128))],512)]),t.footerEnabled?(s(),n("footer",Q,[b(p(h,{appearance:"outline","data-testid":"next-previous-button",onClick:_.goToPrevStep},{default:y(()=>[C(" \u2039 Previous ")]),_:1},8,["onClick"]),[[k,!_.indexCanReverse]]),b(p(h,{disabled:t.nextDisabled,appearance:"primary","data-testid":"next-step-button",onClick:_.goToNextStep},{default:y(()=>[C(" Next \u203A ")]),_:1},8,["disabled","onClick"]),[[k,!_.indexCanAdvance]])])):a("",!0)]),o("aside",j,[o("div",O,[(s(!0),n(v,null,S(t.sidebarContent,(i,l)=>(s(),n("div",{key:i.name,class:m(["wizard-steps__sidebar__item",`wizard-steps__sidebar__item--${l}`])},[d(e.$slots,i.name,{},void 0,!0)],2))),128))])])])}const ie=g(R,[["render",U],["__scopeId","data-v-328c748a"]]);const G={},H={class:"card-icon icon-success mb-3",role:"img"};function J(e,c){return s(),n("i",H," \u2713 ")}const M=g(G,[["render",J],["__scopeId","data-v-f2914797"]]);const W={name:"EntityScanner",components:{IconSuccess:M,KEmptyState:z,KIcon:E},props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},shouldStart:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},emits:["hide-siblings"],data(){return{i:0,isRunning:!1,isComplete:!1,intervalId:null}},watch:{shouldStart(e,c){e!==c&&e===!0&&this.runScanner()}},mounted(){this.shouldStart===!0&&this.runScanner()},beforeUnmount(){clearInterval(this.intervalId)},methods:{runScanner(){this.isRunning=!0,this.isComplete=!1,this.intervalId=setInterval(()=>{this.i++,this.loaderFunction(),(this.i===this.retries||this.canComplete===!0)&&(clearInterval(this.intervalId),this.isRunning=!1,this.isComplete=!0,this.$emit("hide-siblings",!0))},this.interval)}}},X={key:0,class:"scanner"},Y={class:"scanner-content"},Z={key:0,class:"card-icon mb-3"},ee={key:1,class:"card-icon mb-3"},te={key:3},se={key:1};function ne(e,c,t,u,r,_){const h=f("KIcon"),i=f("IconSuccess"),l=f("KEmptyState");return t.shouldStart?(s(),n("div",X,[o("div",Y,[p(l,{"cta-is-hidden":""},{title:y(()=>[r.isRunning?(s(),n("div",Z,[p(h,{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"})])):a("",!0),r.isComplete&&t.hasError===!1&&r.isRunning===!1?(s(),n("div",ee,[p(i)])):a("",!0),r.isRunning?d(e.$slots,"loading-title",{key:2},void 0,!0):a("",!0),r.isRunning===!1?(s(),n("div",te,[t.hasError?d(e.$slots,"error-title",{key:0},void 0,!0):a("",!0),r.isComplete&&t.hasError===!1?d(e.$slots,"complete-title",{key:1},void 0,!0):a("",!0)])):a("",!0)]),message:y(()=>[r.isRunning?d(e.$slots,"loading-content",{key:0},void 0,!0):a("",!0),r.isRunning===!1?(s(),n("div",se,[t.hasError?d(e.$slots,"error-content",{key:0},void 0,!0):a("",!0),r.isComplete&&t.hasError===!1?d(e.$slots,"complete-content",{key:1},void 0,!0):a("",!0)])):a("",!0)]),_:3})])])):a("",!0)}const oe=g(W,[["render",ne],["__scopeId","data-v-ea480f76"]]);export{oe as E,re as F,ie as S};
