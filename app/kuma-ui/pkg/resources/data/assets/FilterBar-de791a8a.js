var oe=Object.defineProperty;var ie=(s,o,t)=>o in s?oe(s,o,{enumerable:!0,configurable:!0,writable:!0,value:t}):s[o]=t;var E=(s,o,t)=>(ie(s,typeof o!="symbol"?o+"":o,t),t);import{K as F}from"./index-52545d1d.js";import{d as W,l as le,a as U,o as p,b as K,a0 as re,w as k,r as Z,e as x,f as m,t as b,c as g,F as L,q as f,p as v,B as J,s as D,W as ue,_ as Y,m as C,N as B,aa as V,at as ce,au as de,av as fe,n as O,aw as pe,ax as me,I as ge,T as he,z as ye,A as ve}from"./index-0447e8bf.js";import{d as be,e as _e,g as ke,b as Se,c as xe}from"./index-12ba49aa.js";import{A as Te}from"./AppCollection-7fc93039.js";import{S as Ce}from"./StatusBadge-46ff6d43.js";import{T as we}from"./TagList-5b11f53f.js";import{_ as Ie}from"./WarningIcon.vue_vue_type_script_setup_true_lang-1d5e83d7.js";const Le={key:0},Fe=W({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},isSelectedRow:{type:[Function,null],default:null},summaryRouteName:{},canUseZones:{type:Boolean}},emits:["change"],setup(s,{emit:o}){const{t,formatIsoDate:S}=le(),l=s,n=o;function i(h){return h.map(r=>{var N,I;const{mesh:c,name:y}=r,a=t(`data-planes.type.${be(r)}`),T=_e(r).filter(P=>P.label==="kuma.io/service"),{status:q}=ke(r);let w;(I=(N=r.dataplaneInsight)==null?void 0:N.mTLS)!=null&&I.certificateExpirationTime?w=S(r.dataplaneInsight.mTLS.certificateExpirationTime):w=t("data-planes.components.data-plane-list.certificate.none");const $=Se(r,l.canUseZones),M=xe(r);return{name:y,type:a,mesh:c,services:T,status:q,warnings:{version_mismatch:$.length>0,cert_expired:M},certificate:w}})}return(h,r)=>{const c=U("RouterLink"),y=U("KTooltip");return p(),K(Te,{class:"data-plane-collection","empty-state-message":f(t)("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":f(t)("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":f(t)("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Type",key:"type"},{label:"Services",key:"services"},{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],"page-number":l.pageNumber,"page-size":l.pageSize,total:l.total,items:l.items?i(l.items):void 0,error:l.error,"is-selected-row":l.isSelectedRow,onChange:r[0]||(r[0]=a=>n("change",a))},re({name:k(({row:a})=>[x(c,{class:"name-link",title:a.name,to:{name:l.summaryRouteName,params:{mesh:a.mesh,dataPlane:a.name},query:{page:l.pageNumber,size:l.pageSize}}},{default:k(()=>[m(b(a.name),1)]),_:2},1032,["title","to"])]),services:k(({row:a})=>[a.services.length>0?(p(),K(we,{key:0,tags:a.services,"should-truncate":"","hide-label-key":""},null,8,["tags"])):(p(),g(L,{key:1},[m(b(f(t)("common.collection.none")),1)],64))]),status:k(({row:a})=>[x(Ce,{status:a.status},null,8,["status"])]),warnings:k(({row:a})=>[Object.values(a.warnings).some(_=>_)?(p(),K(y,{key:0},{content:k(()=>[v("ul",null,[(p(!0),g(L,null,J(a.warnings,(_,T)=>(p(),g(L,{key:T},[_?(p(),g("li",Le,b(f(t)(`data-planes.components.data-plane-list.${T}`)),1)):D("",!0)],64))),128))])]),default:k(()=>[m(),x(Ie,{class:"mr-1",size:f(F),"hide-title":""},null,8,["size"])]),_:2},1024)):(p(),g(L,{key:1},[m(b(f(t)("common.collection.none")),1)],64))]),details:k(({row:a})=>[x(c,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:a.name}}},{default:k(()=>[m(b(f(t)("common.collection.details_link"))+" ",1),x(f(ue),{display:"inline-block",decorative:"",size:f(F)},null,8,["size"])]),_:2},1032,["to"])]),_:2},[h.$slots.toolbar?{name:"toolbar",fn:k(()=>[Z(h.$slots,"toolbar",{},void 0,!0)]),key:"0"}:void 0]),1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","page-number","page-size","total","items","error","is-selected-row"])}}});const nt=Y(Fe,[["__scopeId","data-v-9d0641dc"]]);function Ne(s,o,t){return Math.max(o,Math.min(s,t))}const Ae=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class ze{constructor(o,t){E(this,"commands");E(this,"keyMap");E(this,"boundTriggerShortcuts");this.commands=t,this.keyMap=Object.fromEntries(Object.entries(o).map(([S,l])=>[S.toLowerCase(),l])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(o){Ee(o,this.keyMap,this.commands)}}function Ee(s,o,t){const S=Be(s.code),l=[s.ctrlKey?"ctrl":"",s.shiftKey?"shift":"",s.altKey?"alt":"",S].filter(h=>h!=="").join("+"),n=o[l];if(!n)return;const i=t[n];i.isAllowedContext&&!i.isAllowedContext(s)||(i.shouldPreventDefaultAction&&s.preventDefault(),!(i.isDisabled&&i.isDisabled())&&i.trigger(s))}function Be(s){return Ae.includes(s)?"":s.replace(/^Key/,"").toLowerCase()}function De(s,o){const t=" "+s,S=t.matchAll(/ ([-\s\w]+):\s*/g),l=[];for(const n of Array.from(S)){if(n.index===void 0)continue;const i=qe(n[1]);if(o.length>0&&!o.includes(i))throw new Error(`Unknown field “${i}”. Known fields: ${o.join(", ")}`);const h=n.index+n[0].length,r=t.substring(h);let c;if(/^\s*["']/.test(r)){const a=r.match(/['"](.*?)['"]/);if(a!==null)c=a[1];else throw new Error(`Quote mismatch for field “${i}”.`)}else{const a=r.indexOf(" "),_=a===-1?r.length:a;c=r.substring(0,_)}c!==""&&l.push([i,c])}return l}function qe(s){return s.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(o,t)=>t===0?o:o.substring(1).toUpperCase())}let H=0;const $e=(s="unique")=>(H++,`${s}-${H}`),G=s=>(ye("data-v-349996e3"),s=s(),ve(),s),Me=G(()=>v("span",{class:"visually-hidden"},"Focus filter",-1)),Pe={class:"k-filter-icon"},je=["for"],Ke=["id","placeholder"],Re={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},Qe={class:"k-suggestion-list"},Ue={key:0,class:"k-filter-bar-error"},Ve={key:0},Oe=["title","data-filter-field"],He={class:"visually-hidden"},We=G(()=>v("span",{class:"visually-hidden"},"Clear query",-1)),Ze=W({__name:"FilterBar",props:{id:{type:String,required:!1,default:()=>$e("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(s,{emit:o}){const t=s,S=o,l=C(null),n=C(null),i=C(t.query),h=C([]),r=C(null),c=C(!1),y=C(-1),a=B(()=>Object.keys(t.fields)),_=B(()=>Object.entries(t.fields).slice(0,5).map(([e,u])=>({fieldName:e,...u}))),T=B(()=>a.value.length>0?`Filter by ${a.value.join(", ")}`:"Filter"),q=B(()=>t.placeholder??T.value);V(()=>h.value,function(e,u){ne(e,u)||(r.value=null,S("fields-change",{fields:e,query:i.value}))}),V(()=>i.value,function(){i.value===""&&(r.value=null),c.value=!0});const w={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},$={submitQuery:{trigger:I,isAllowedContext(e){return n.value!==null&&e.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:P,isAllowedContext(e){return n.value!==null&&e.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:X,isAllowedContext(e){return n.value!==null&&e.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:j,isAllowedContext(e){return l.value!==null&&e.composedPath().includes(l.value)}}};function M(){const e=new ze(w,$);ge(function(){e.registerListener()}),he(function(){e.unRegisterListener()}),A(i.value)}M();function N(e){const u=e.target;A(u.value)}function I(){if(n.value instanceof HTMLInputElement)if(y.value===-1)A(n.value.value),c.value=!1;else{const e=_.value[y.value].fieldName;e&&Q(n.value,e)}}function P(){R(1)}function X(){R(-1)}function R(e){y.value=Ne(y.value+e,-1,_.value.length-1)}function ee(){n.value instanceof HTMLInputElement&&n.value.focus()}function te(e){const d=e.currentTarget.getAttribute("data-filter-field");d&&n.value instanceof HTMLInputElement&&Q(n.value,d)}function Q(e,u){const d=i.value===""||i.value.endsWith(" ")?"":" ";i.value+=d+u+":",e.focus(),y.value=-1}function se(){i.value="",n.value instanceof HTMLInputElement&&(n.value.value="",n.value.focus(),A(""))}function ae(e){e.relatedTarget===null&&j(),l.value instanceof HTMLElement&&e.relatedTarget instanceof Node&&!l.value.contains(e.relatedTarget)&&j()}function j(){c.value=!1}function A(e){r.value=null;try{const u=De(e,a.value);u.sort((d,z)=>d[0].localeCompare(z[0])),h.value=u}catch(u){if(u instanceof Error)r.value=u,c.value=!0;else throw u}}function ne(e,u){return JSON.stringify(e)===JSON.stringify(u)}return(e,u)=>(p(),g("div",{ref_key:"filterBar",ref:l,class:"k-filter-bar","data-testid":"k-filter-bar"},[v("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:ee},[Me,m(),v("span",Pe,[x(f(ce),{decorative:"","data-testid":"k-filter-bar-filter-icon","hide-title":"",size:f(F)},null,8,["size"])])]),m(),v("label",{for:`${t.id}-filter-bar-input`,class:"visually-hidden"},[Z(e.$slots,"default",{},()=>[m(b(T.value),1)],!0)],8,je),m(),de(v("input",{id:`${t.id}-filter-bar-input`,ref_key:"filterInput",ref:n,"onUpdate:modelValue":u[0]||(u[0]=d=>i.value=d),class:"k-filter-bar-input",type:"text",placeholder:q.value,"data-testid":"k-filter-bar-filter-input",onFocus:u[1]||(u[1]=d=>c.value=!0),onBlur:ae,onChange:N},null,40,Ke),[[fe,i.value]]),m(),c.value?(p(),g("div",Re,[v("div",Qe,[r.value!==null?(p(),g("p",Ue,b(r.value.message),1)):(p(),g("button",{key:1,class:O(["k-submit-query-button",{"k-submit-query-button-is-selected":y.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:I},`
          Submit `+b(i.value),3)),m(),(p(!0),g(L,null,J(_.value,(d,z)=>(p(),g("div",{key:`${t.id}-${z}`,class:O(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":y.value===z}])},[v("b",null,b(d.fieldName),1),d.description!==""?(p(),g("span",Ve,": "+b(d.description),1)):D("",!0),m(),v("button",{class:"k-apply-suggestion-button",title:`Add ${d.fieldName}:`,type:"button","data-filter-field":d.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:te},[v("span",He,"Add "+b(d.fieldName)+":",1),m(),x(f(pe),{decorative:"","hide-title":"",size:f(F)},null,8,["size"])],8,Oe)],2))),128))])])):D("",!0),m(),i.value!==""?(p(),g("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:se},[We,m(),x(f(me),{decorative:"","hide-title":"",size:f(F)},null,8,["size"])])):D("",!0)],512))}});const ot=Y(Ze,[["__scopeId","data-v-349996e3"]]);export{nt as D,ot as F};
