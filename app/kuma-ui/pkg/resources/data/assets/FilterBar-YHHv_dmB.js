var Y=Object.defineProperty;var Z=(t,o,n)=>o in t?Y(t,o,{enumerable:!0,configurable:!0,writable:!0,value:n}):t[o]=n;var w=(t,o,n)=>(Z(t,typeof o!="symbol"?o+"":o,n),n);import{d as G,ac as X,C as y,M as T,a6 as I,o as v,c as m,m as f,f as d,e as q,p as k,K as A,ao as ee,r as te,t as S,ap as ne,aq as se,n as B,F as ie,H as oe,q as M,ar as le,as as re,at as ue,au as ae,x as ce,y as de,_ as fe}from"./index-CP9JG8i6.js";function ge(t,o,n){return Math.max(o,Math.min(t,n))}const pe=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class ve{constructor(o,n){w(this,"commands");w(this,"keyMap");w(this,"boundTriggerShortcuts");this.commands=n,this.keyMap=Object.fromEntries(Object.entries(o).map(([h,u])=>[h.toLowerCase(),u])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(o){me(o,this.keyMap,this.commands)}}function me(t,o,n){const h=he(t.code),u=[t.ctrlKey?"ctrl":"",t.shiftKey?"shift":"",t.altKey?"alt":"",h].filter(b=>b!=="").join("+"),i=o[u];if(!i)return;const s=n[i];s.isAllowedContext&&!s.isAllowedContext(t)||(s.shouldPreventDefaultAction&&t.preventDefault(),!(s.isDisabled&&s.isDisabled())&&s.trigger(t))}function he(t){return pe.includes(t)?"":t.replace(/^Key/,"").toLowerCase()}function be(t,o){const n=" "+t,h=n.matchAll(/ ([-\s\w]+):\s*/g),u=[];for(const i of Array.from(h)){if(i.index===void 0)continue;const s=ye(i[1]);if(o.length>0&&!o.includes(s))throw new Error(`Unknown field “${s}”. Known fields: ${o.join(", ")}`);const b=i.index+i[0].length,a=n.substring(b);let c;if(/^\s*["']/.test(a)){const g=a.match(/['"](.*?)['"]/);if(g!==null)c=g[1];else throw new Error(`Quote mismatch for field “${s}”.`)}else{const g=a.indexOf(" "),_=g===-1?a.length:g;c=a.substring(0,_)}c!==""&&u.push([s,c])}return u}function ye(t){return t.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(o,n)=>n===0?o:o.substring(1).toUpperCase())}const j=t=>(ce("data-v-e7ca9a15"),t=t(),de(),t),ke=j(()=>f("span",{class:"visually-hidden"},"Focus filter",-1)),Se={class:"k-filter-icon"},_e=["for"],Ce=["id","placeholder"],xe={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},we={class:"k-suggestion-list"},Te={key:0,class:"k-filter-bar-error"},Fe={key:0},Ie=["title","data-filter-field"],qe={class:"visually-hidden"},Ae=j(()=>f("span",{class:"visually-hidden"},"Clear query",-1)),Me=G({__name:"FilterBar",props:{id:{type:String,required:!1,default:()=>X("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(t,{emit:o}){const n=t,h=o,u=y(null),i=y(null),s=y(n.query),b=y([]),a=y(null),c=y(!1),p=y(-1),g=T(()=>Object.keys(n.fields)),_=T(()=>Object.entries(n.fields).slice(0,5).map(([e,l])=>({fieldName:e,...l}))),N=T(()=>g.value.length>0?`Filter by ${g.value.join(", ")}`:"Filter"),O=T(()=>n.placeholder??N.value);I(()=>b.value,function(e,l){W(e,l)||(a.value=null,h("fields-change",{fields:e,query:s.value}))}),I(()=>n.query,()=>{s.value=n.query,C(s.value)},{immediate:!0}),I(()=>s.value,function(){s.value===""&&(a.value=null)});const D={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},z={submitQuery:{trigger:E,isAllowedContext(e){return i.value!==null&&e.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:K,isAllowedContext(e){return i.value!==null&&e.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:V,isAllowedContext(e){return i.value!==null&&e.composedPath().includes(i.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:F,isAllowedContext(e){return u.value!==null&&e.composedPath().includes(u.value)}}};function Q(){const e=new ve(D,z);ue(function(){e.registerListener()}),ae(function(){e.unRegisterListener()}),C(s.value)}Q();function $(e){const l=e.target;C(l.value)}function E(){if(i.value instanceof HTMLInputElement)if(p.value===-1)C(i.value.value),c.value=!1;else{const e=_.value[p.value].fieldName;e&&P(i.value,e)}}function K(){L(1)}function V(){L(-1)}function L(e){p.value=ge(p.value+e,-1,_.value.length-1)}function H(){i.value instanceof HTMLInputElement&&i.value.focus()}function U(e){const r=e.currentTarget.getAttribute("data-filter-field");r&&i.value instanceof HTMLInputElement&&P(i.value,r)}function P(e,l){const r=s.value===""||s.value.endsWith(" ")?"":" ";s.value+=r+l+":",e.focus(),p.value=-1}function R(){s.value="",i.value instanceof HTMLInputElement&&(i.value.value="",i.value.focus(),C(""))}function J(e){e.relatedTarget===null&&F(),u.value instanceof HTMLElement&&e.relatedTarget instanceof Node&&!u.value.contains(e.relatedTarget)&&F()}function F(){c.value=!1}function C(e){a.value=null;try{const l=be(e,g.value);l.sort((r,x)=>r[0].localeCompare(x[0])),b.value=l}catch(l){if(l instanceof Error)a.value=l,c.value=!0;else throw l}}function W(e,l){return JSON.stringify(e)===JSON.stringify(l)}return(e,l)=>(v(),m("div",{ref_key:"filterBar",ref:u,class:"k-filter-bar","data-testid":"k-filter-bar"},[f("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:H},[ke,d(),f("span",Se,[q(k(ee),{decorative:"","data-testid":"k-filter-bar-filter-icon",size:k(A)},null,8,["size"])])]),d(),f("label",{for:`${n.id}-filter-bar-input`,class:"visually-hidden"},[te(e.$slots,"default",{},()=>[d(S(N.value),1)],!0)],8,_e),d(),ne(f("input",{id:`${n.id}-filter-bar-input`,ref_key:"filterInput",ref:i,"onUpdate:modelValue":l[0]||(l[0]=r=>s.value=r),class:"k-filter-bar-input",type:"text",placeholder:O.value,"data-testid":"k-filter-bar-filter-input",onFocus:l[1]||(l[1]=r=>c.value=!0),onBlur:J,onChange:$},null,40,Ce),[[se,s.value]]),d(),c.value?(v(),m("div",xe,[f("div",we,[a.value!==null?(v(),m("p",Te,S(a.value.message),1)):(v(),m("button",{key:1,class:B(["k-submit-query-button",{"k-submit-query-button-is-selected":p.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:E},`
          Submit `+S(s.value),3)),d(),(v(!0),m(ie,null,oe(_.value,(r,x)=>(v(),m("div",{key:`${n.id}-${x}`,class:B(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":p.value===x}])},[f("b",null,S(r.fieldName),1),r.description!==""?(v(),m("span",Fe,": "+S(r.description),1)):M("",!0),d(),f("button",{class:"k-apply-suggestion-button",title:`Add ${r.fieldName}:`,type:"button","data-filter-field":r.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:U},[f("span",qe,"Add "+S(r.fieldName)+":",1),d(),q(k(le),{decorative:"",size:k(A)},null,8,["size"])],8,Ie)],2))),128))])])):M("",!0),d(),s.value!==""?(v(),m("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:R},[Ae,d(),q(k(re),{decorative:"",size:k(A)},null,8,["size"])])):M("",!0)],512))}}),Le=fe(Me,[["__scopeId","data-v-e7ca9a15"]]);export{Le as F};
