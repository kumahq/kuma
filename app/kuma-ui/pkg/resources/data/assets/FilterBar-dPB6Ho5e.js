var V=Object.defineProperty;var U=(t,i,r)=>i in t?V(t,i,{enumerable:!0,configurable:!0,writable:!0,value:r}):t[i]=r;var C=(t,i,r)=>U(t,typeof i!="symbol"?i+"":i,r);import{d as D,ad as A,T as k,o as g,m as Z,w as Q,c as h,ae as W,K as Y,ab as G,y,af as J,ag as X,ah as ee,s as c,e as b,b as B,p as L,ai as q,aj as te,a as ie,t as w,A as re,ak as se,n as M,F as ne,v as oe,q as _,P as ae,_ as le}from"./index-DIs9RbIP.js";const ue=["aria-hidden"],z='<path d="M9.4 18L8 16.6L12.6 12L8 7.4L9.4 6L15.4 12L9.4 18Z" fill="currentColor"/>',de=D({__name:"ChevronRightIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:A,validator:t=>{if(typeof t=="number"&&t>0)return!0;if(typeof t=="string"){const i=String(t).replace(/px/gi,""),r=Number(i);if(r&&!isNaN(r)&&Number.isInteger(r)&&r>0)return!0}return!1}},as:{type:String,required:!1,default:"span"},staticIds:{type:Boolean,default:!1}},setup(t){const i=t,r=k(()=>{if(typeof i.size=="number"&&i.size>0)return`${i.size}px`;if(typeof i.size=="string"){const o=String(i.size).replace(/px/gi,""),a=Number(o);if(a&&!isNaN(a)&&Number.isInteger(a)&&a>0)return`${a}px`}return A}),p=k(()=>({boxSizing:"border-box",color:i.color,display:i.display,flexShrink:"0",height:r.value,lineHeight:"0",width:r.value,pointerEvents:i.decorative?"none":void 0})),m=o=>{const a={},f=Math.random().toString(36).substring(2,12);return o.replace(/id="([^"]+)"/g,(d,l)=>{const x=`${f}-${l}`;return a[l]=x,`id="${x}"`}).replace(/#([^\s^")]+)/g,(d,l)=>a[l]?`#${a[l]}`:d)},S={"<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;","&":"&amp;"},u=o=>o.replace(/[<>"'&]/g,a=>S[a]),v=`${i.title?`<title data-testid="kui-icon-svg-title">${u(i.title)}</title>`:""}${i.staticIds?z:m(z)}`;return(o,a)=>(g(),Z(Y(t.as),{"aria-hidden":t.decorative?"true":void 0,class:"kui-icon chevron-right-icon","data-testid":"kui-icon-wrapper-chevron-right-icon",style:W(p.value),tabindex:t.decorative?"-1":void 0},{default:Q(()=>[(g(),h("svg",{"aria-hidden":t.decorative?"true":void 0,"data-testid":"kui-icon-svg-chevron-right-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg",innerHTML:v},null,8,ue))]),_:1},8,["aria-hidden","style","tabindex"]))}}),ce=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class fe{constructor(i,r){C(this,"commands");C(this,"keyMap");C(this,"boundTriggerShortcuts");this.commands=r,this.keyMap=Object.fromEntries(Object.entries(i).map(([p,m])=>[p.toLowerCase(),m])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(i){ge(i,this.keyMap,this.commands)}}function ge(t,i,r){const p=pe(t.code),m=[t.ctrlKey?"ctrl":"",t.shiftKey?"shift":"",t.altKey?"alt":"",p].filter(v=>v!=="").join("+"),S=i[m];if(!S)return;const u=r[S];u.isAllowedContext&&!u.isAllowedContext(t)||(u.shouldPreventDefaultAction&&t.preventDefault(),!(u.isDisabled&&u.isDisabled())&&u.trigger(t))}function pe(t=""){return ce.includes(t)?"":t.replace(/^Key/,"").toLowerCase()}const me={class:"filter-bar-icon"},ve=["for"],be=["id","placeholder"],he={key:0,class:"suggestion-box","data-testid":"filter-bar-suggestion-box"},ye={class:"suggestion-list"},Se={key:0,class:"filter-bar-error"},we={key:0},ke=["title","data-filter-field"],xe={class:"visually-hidden"},Ce=D({__name:"FilterBar",props:{fields:{},placeholder:{default:""},query:{default:""},id:{default:()=>G("filter-bar")}},emits:["change"],setup(t,{emit:i}){const r=t,p=y(),m=i,S=s=>{s!=null&&s.target&&(m("change",new FormData(s.target)),f.value=!1)},u=s=>{m("change",new FormData(p.value))},v=y(null),o=y(null),a=y(null),f=y(!1),d=y(r.query);J(()=>r.query,s=>{d.value=s});const l=y(0),x=k(()=>Object.keys(r.fields)),T=k(()=>Object.entries(r.fields).slice(0,5).map(([s,e])=>({fieldName:s,...e}))),F=k(()=>x.value.length>0?`Filter by ${x.value.join(", ")}`:"Filter"),j=k(()=>r.placeholder??F.value),P={ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},E={jumpToNextSuggestion:{trigger:()=>$(1),isAllowedContext(s){return o.value!==null&&s.composedPath().includes(o.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:()=>$(-1),isAllowedContext(s){return o.value!==null&&s.composedPath().includes(o.value)},shouldPreventDefaultAction:!0}},N=new fe(P,E);X(function(){N.registerListener()}),ee(function(){N.unRegisterListener()});function $(s){const e=T.value.length;let n=l.value+s;n===-1&&(n=e),l.value=n%(e+1)}function K(){o.value instanceof HTMLInputElement&&o.value.focus()}function O(s){const n=s.currentTarget.getAttribute("data-filter-field");n&&o.value instanceof HTMLInputElement&&R(o.value,n)}function R(s,e){const n=d.value===""||d.value.endsWith(" ")?"":" ";d.value+=n+e+":",s.focus(),l.value=0}function H(s){s.relatedTarget===null&&(f.value=!1),v.value instanceof HTMLElement&&s.relatedTarget instanceof Node&&!v.value.contains(s.relatedTarget)&&(f.value=!1)}return(s,e)=>(g(),h("div",{ref_key:"filterBar",ref:v,class:"filter-bar","data-testid":"filter-bar"},[c("form",{ref_key:"$form",ref:p,onSubmit:ae(S,["prevent"])},[c("button",{class:"focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"filter-bar-focus-filter-input-button",onClick:K},[e[4]||(e[4]=c("span",{class:"visually-hidden"},"Focus filter",-1)),e[5]||(e[5]=b()),c("span",me,[B(L(te),{decorative:"","data-testid":"filter-bar-filter-icon","hide-title":"",size:L(q)},null,8,["size"])])]),e[9]||(e[9]=b()),c("label",{for:`${r.id}-filter-bar-input`,class:"visually-hidden"},[ie(s.$slots,"default",{},()=>[b(w(F.value),1)],!0)],8,ve),e[10]||(e[10]=b()),re(c("input",{id:`${r.id}-filter-bar-input`,ref_key:"filterInput",ref:o,"onUpdate:modelValue":e[0]||(e[0]=n=>d.value=n),class:"filter-bar-input",type:"search",placeholder:j.value,"data-testid":"filter-bar-filter-input",name:"s",onFocus:e[1]||(e[1]=n=>f.value=!0),onInput:e[2]||(e[2]=n=>f.value=!0),onBlur:H,onSearch:e[3]||(e[3]=n=>{n.target.value.length===0&&(u(),f.value=!0)})},null,40,be),[[se,d.value]]),e[11]||(e[11]=b()),f.value?(g(),h("div",he,[c("div",ye,[a.value!==null?(g(),h("p",Se,w(a.value.message),1)):(g(),h("button",{key:1,type:"submit",class:M(["submit-query-button",{"submit-query-button-is-selected":l.value===0}]),"data-testid":"filter-bar-submit-query-button"},`
            Submit `+w(d.value),3)),e[8]||(e[8]=b()),(g(!0),h(ne,null,oe(T.value,(n,I)=>(g(),h("div",{key:`${r.id}-${I}`,class:M(["suggestion-list-item",{"suggestion-list-item-is-selected":l.value===I+1}])},[c("b",null,w(n.fieldName),1),n.description!==""?(g(),h("span",we,": "+w(n.description),1)):_("",!0),e[7]||(e[7]=b()),c("button",{class:"apply-suggestion-button",title:`Add ${n.fieldName}:`,type:"button","data-filter-field":n.fieldName,"data-testid":"filter-bar-apply-suggestion-button",onClick:O},[c("span",xe,"Add "+w(n.fieldName)+":",1),e[6]||(e[6]=b()),B(L(de),{decorative:"","hide-title":"",size:L(q)},null,8,["size"])],8,ke)],2))),128))])])):_("",!0)],544)],512))}}),Te=le(Ce,[["__scopeId","data-v-7603a67a"]]);export{Te as F};
