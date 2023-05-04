import{d as B,j as D,$ as F}from"./kongponents.es-ba82ceca.js";import{_ as w}from"./_plugin-vue_export-helper-c27b6911.js";import{o as t,j as s,d as E,r as y,v as L,k as A,a1 as P,i as u,g as h,w as $,b as p,f as o,h as r,J as f,t as R,Y as k,c as I,F as z,q as C,$ as q,an as x}from"./index-bd38c154.js";import{Q as N}from"./QueryParameter-70743f73.js";const V={},j={class:"icon-success mb-3",role:"img"};function T(a,c){return t(),s("i",j,`
    ✓
  `)}const Q=w(V,[["render",T],["__scopeId","data-v-fdd227f8"]]),J={key:0,class:"scanner"},M={class:"scanner-content"},U={key:0,class:"mb-3"},Y={key:1,class:"mb-3"},G={key:3},H={key:1},K=E({__name:"EntityScanner",props:{interval:{type:Number,required:!1,default:1e3},retries:{type:Number,required:!1,default:3600},shouldStart:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},loaderFunction:{type:Function,required:!0},canComplete:{type:Boolean,default:!1}},emits:["hide-siblings"],setup(a,{emit:c}){const e=a,n=y(0),l=y(!1),v=y(!1),g=y(null);L(()=>e.shouldStart,function(i,d){i!==d&&i===!0&&S()}),A(function(){e.shouldStart===!0&&S()}),P(function(){b()});function S(){l.value=!0,v.value=!1,b(),g.value=window.setInterval(()=>{n.value++,e.loaderFunction(),(n.value===e.retries||e.canComplete===!0)&&(b(),l.value=!1,v.value=!0,c("hide-siblings",!0))},e.interval)}function b(){g.value!==null&&window.clearInterval(g.value)}return(i,d)=>a.shouldStart?(t(),s("div",J,[u("div",M,[h(p(D),{"cta-is-hidden":""},{title:$(()=>[l.value?(t(),s("div",U,[h(p(B),{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"})])):o("",!0),r(),v.value&&a.hasError===!1&&l.value===!1?(t(),s("div",Y,[h(Q)])):o("",!0),r(),l.value?f(i.$slots,"loading-title",{key:2},void 0,!0):o("",!0),r(),l.value===!1?(t(),s("div",G,[a.hasError?f(i.$slots,"error-title",{key:0},void 0,!0):o("",!0),r(),v.value&&a.hasError===!1?f(i.$slots,"complete-title",{key:1},void 0,!0):o("",!0)])):o("",!0)]),message:$(()=>[l.value?f(i.$slots,"loading-content",{key:0},void 0,!0):o("",!0),r(),l.value===!1?(t(),s("div",H,[a.hasError?f(i.$slots,"error-content",{key:0},void 0,!0):o("",!0),r(),v.value&&a.hasError===!1?f(i.$slots,"complete-content",{key:1},void 0,!0):o("",!0)])):o("",!0)]),_:3})])])):o("",!0)}});const he=w(K,[["__scopeId","data-v-d6fe0c46"]]),O={class:"form-line-wrapper"},W={key:0,class:"form-line__col"},X=["for"],Z=E({__name:"FormFragment",props:{title:{type:String,required:!1,default:null},forAttr:{type:String,required:!1,default:null},allInline:{type:Boolean,default:!1},hideLabelCol:{type:Boolean,default:!1},equalCols:{type:Boolean,default:!1},shiftRight:{type:Boolean,default:!1}},setup(a){const c=a;return(e,n)=>(t(),s("div",O,[u("div",{class:k(["form-line",{"has-equal-cols":c.equalCols}])},[c.hideLabelCol?o("",!0):(t(),s("div",W,[u("label",{for:c.forAttr,class:"k-input-label"},R(c.title)+`:
        `,9,X)])),r(),u("div",{class:k(["form-line__col",{"is-inline":c.allInline,"is-shifted-right":c.shiftRight}])},[f(e.$slots,"default")],2)],2)]))}});const ye=w(Z,[["__scopeId","data-v-aa1ca9d8"]]),ee={class:"wizard-steps"},te={class:"wizard-steps__content-wrapper"},se={class:"wizard-steps__indicator"},ae={class:"wizard-steps__indicator__controls",role:"tablist","aria-label":"steptabs"},ne=["aria-selected","aria-controls"],oe={class:"wizard-steps__content"},le={ref:"wizardForm",autocomplete:"off"},re=["id","aria-labelledby"],ie={key:0,class:"wizard-steps__footer"},de={class:"wizard-steps__sidebar"},ue={class:"wizard-steps__sidebar__content"},ce=E({__name:"StepSkeleton",props:{steps:{type:Array,required:!0},sidebarContent:{type:Array,required:!0},footerEnabled:{type:Boolean,default:!0},nextDisabled:{type:Boolean,default:!0}},emits:["go-to-step"],setup(a,{emit:c}){const e=a,n=y(0),l=y(null),v=I(()=>n.value>=e.steps.length-1),g=I(()=>n.value<=0);A(function(){const d=N.get("step");n.value=d?parseInt(d):0,n.value in e.steps&&(l.value=e.steps[n.value].slug)});function S(){n.value++,i(n.value)}function b(){n.value--,i(n.value)}function i(d){l.value=e.steps[d].slug,N.set("step",d),c("go-to-step",d)}return(d,_e)=>(t(),s("div",ee,[u("div",te,[u("header",se,[u("ul",ae,[(t(!0),s(z,null,C(a.steps,(_,m)=>(t(),s("li",{key:_.slug,"aria-selected":l.value===_.slug?"true":"false","aria-controls":`wizard-steps__content__item--${m}`,class:k([{"is-complete":m<=n.value},"wizard-steps__indicator__item"])},[u("span",null,R(_.label),1)],10,ne))),128))])]),r(),u("div",oe,[u("form",le,[(t(!0),s(z,null,C(a.steps,(_,m)=>(t(),s("div",{id:`wizard-steps__content__item--${m}`,key:_.slug,"aria-labelledby":`wizard-steps__content__item--${m}`,role:"tabpanel",tabindex:"0",class:"wizard-steps__content__item"},[l.value===_.slug?f(d.$slots,_.slug,{key:0},void 0,!0):o("",!0)],8,re))),128))],512)]),r(),e.footerEnabled?(t(),s("footer",ie,[q(h(p(F),{appearance:"outline","data-testid":"next-previous-button",onClick:b},{default:$(()=>[h(p(B),{icon:"chevronLeft",color:"currentColor",size:"16","hide-title":""}),r(`

          Previous
        `)]),_:1},512),[[x,!p(g)]]),r(),q(h(p(F),{disabled:e.nextDisabled,appearance:"primary","data-testid":"next-step-button",onClick:S},{default:$(()=>[r(`
          Next

          `),h(p(B),{icon:"chevronRight",color:"currentColor",size:"16","hide-title":""})]),_:1},8,["disabled"]),[[x,!p(v)]])])):o("",!0)]),r(),u("aside",de,[u("div",ue,[(t(!0),s(z,null,C(e.sidebarContent,(_,m)=>(t(),s("div",{key:_.name,class:k(["wizard-steps__sidebar__item",`wizard-steps__sidebar__item--${m}`])},[f(d.$slots,_.name,{},void 0,!0)],2))),128))])])]))}});const ge=w(ce,[["__scopeId","data-v-4d0f6a65"]]);export{he as E,ye as F,ge as S};
